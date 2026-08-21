#!/usr/bin/env bash
# Asserts live infrastructure against Terraform outputs, never hardcoded names.

failures=0

check() {
  local name=$1 expected=$2 actual=$3
  if [ "$actual" = "$expected" ]; then
    printf 'ok    %s\n' "$name"
  else
    printf 'FAIL  %s: expected %q, got %q\n' "$name" "$expected" "$actual"
    failures=$((failures + 1))
  fi
}

fail() {
  printf 'FAIL  %s\n' "$1"
  failures=$((failures + 1))
}

terraform_dir=$(cd "$(dirname "$0")/../terraform" && pwd)
outputs=$(terraform -chdir="$terraform_dir" output -json) || {
  echo "cannot read terraform outputs; run terraform apply first" >&2
  exit 1
}

get_output() { jq -r ".$1.value" <<<"$outputs"; }

bucket=$(get_output bucket_name)
table=$(get_output table_name)
queue_url=$(get_output queue_url)
queue_arn=$(get_output queue_arn)
dlq_url=$(get_output dlq_url)
api_role=$(get_output api_role_arn)
worker_role=$(get_output worker_role_arn)

echo "== S3 =="
aws s3api head-bucket --bucket "$bucket" >/dev/null 2>&1 \
  && echo "ok    bucket $bucket exists" || fail "bucket $bucket missing"

pab=$(aws s3api get-public-access-block --bucket "$bucket" 2>/dev/null \
  | jq -r '.PublicAccessBlockConfiguration | [.BlockPublicAcls, .BlockPublicPolicy, .IgnorePublicAcls, .RestrictPublicBuckets] | join(",")')
check "public access fully blocked" "true,true,true,true" "$pab"

echo "== DynamoDB =="
desc=$(aws dynamodb describe-table --table-name "$table" 2>/dev/null)
check "table status" "ACTIVE" "$(jq -r '.Table.TableStatus' <<<"$desc")"
check "key schema" "PK:HASH" "$(jq -r '.Table.KeySchema | map(.AttributeName + ":" + .KeyType) | join(",")' <<<"$desc")"
check "no secondary indexes" "0" "$(jq -r '(.Table.GlobalSecondaryIndexes // []) + (.Table.LocalSecondaryIndexes // []) | length' <<<"$desc")"

echo "== SQS =="
attrs=$(aws sqs get-queue-attributes --queue-url "$queue_url" --attribute-names All 2>/dev/null)
check "visibility timeout" "180" "$(jq -r '.Attributes.VisibilityTimeout' <<<"$attrs")"
redrive=$(jq -r '.Attributes.RedrivePolicy // empty' <<<"$attrs")
if [ -z "$redrive" ]; then
  fail "queue has no redrive policy"
else
  check "redrive max receive count" "3" "$(jq -r '.maxReceiveCount' <<<"$redrive")"
  check "redrive target" "${dlq_url##*/}" "$(jq -r '.deadLetterTargetArn' <<<"$redrive" | awk -F: '{print $NF}')"
fi

aws sqs get-queue-attributes --queue-url "$dlq_url" --attribute-names QueueArn >/dev/null 2>&1 \
  && echo "ok    dead-letter queue exists" || fail "dead-letter queue missing"

echo "== IAM =="
for role_arn in "$api_role" "$worker_role"; do
  role=${role_arn##*/}
  policies=$(aws iam list-role-policies --role-name "$role" 2>/dev/null | jq -r '.PolicyNames[]')
  if [ -z "$policies" ]; then
    fail "role $role has no inline policy"
    continue
  fi
  for policy in $policies; do
    doc=$(aws iam get-role-policy --role-name "$role" --policy-name "$policy" 2>/dev/null | jq '.PolicyDocument')
    wildcards=$(jq -r '[.Statement[].Resource] | flatten | map(select(. == "*")) | length' <<<"$doc")
    check "$role/$policy has no wildcard resource" "0" "$wildcards"
  done
done

echo "== S3 round trip =="
key="uploads/verify-$$/probe.txt"
probe=$(mktemp)
echo "parcel-verify-$$" >"$probe"
if aws s3api put-object --bucket "$bucket" --key "$key" --body "$probe" >/dev/null 2>&1; then
  roundtrip=$(mktemp)
  aws s3api get-object --bucket "$bucket" --key "$key" "$roundtrip" >/dev/null 2>&1
  check "round-tripped object contents" "$(cat "$probe")" "$(cat "$roundtrip")"
  aws s3api delete-object --bucket "$bucket" --key "$key" >/dev/null 2>&1 \
    && echo "ok    probe object deleted" || fail "could not delete probe object"
  rm -f "$roundtrip"
else
  fail "could not put probe object"
fi
rm -f "$probe"

echo
if [ "$failures" -eq 0 ]; then
  echo "all checks passed"
else
  echo "$failures check(s) failed"
fi
exit $((failures > 0))
