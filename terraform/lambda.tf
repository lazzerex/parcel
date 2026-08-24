data "archive_file" "api" {
  type        = "zip"
  source_dir  = "${path.module}/../api"
  output_path = "${path.module}/build/api.zip"

  excludes = [
    ".venv",
    "tests",
    "__pycache__",
    ".pytest_cache",
    "requirements-dev.txt",
    "pytest.ini",
  ]
}

resource "aws_lambda_function" "api" {
  function_name    = "${var.project_name}-api"
  role             = aws_iam_role.api.arn
  handler          = "handler.handler"
  runtime          = "python3.12"
  timeout          = 10
  filename         = data.archive_file.api.output_path
  source_code_hash = data.archive_file.api.output_base64sha256

  environment {
    variables = {
      DYNAMODB_TABLE = aws_dynamodb_table.metadata.name
      S3_BUCKET      = aws_s3_bucket.files.bucket
      SQS_QUEUE_URL  = aws_sqs_queue.jobs.url
    }
  }
}

# Rebuilds whenever a worker source file changes; local-exec is the simplest
# way to produce a Lambda-ready binary without a separate CI/build pipeline.
resource "null_resource" "build_worker" {
  triggers = {
    source_hash = sha1(join("", [
      for f in fileset("${path.module}/../worker", "**/*.go") :
      filesha1("${path.module}/../worker/${f}")
    ]))
  }

  provisioner "local-exec" {
    command = "mkdir -p ${abspath(path.module)}/build && cd ${path.module}/../worker && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ${abspath(path.module)}/build/bootstrap ./cmd/worker"
  }
}

data "archive_file" "worker" {
  type        = "zip"
  source_file = "${path.module}/build/bootstrap"
  output_path = "${path.module}/build/worker.zip"
  depends_on  = [null_resource.build_worker]
}

resource "aws_lambda_function" "worker" {
  function_name    = "${var.project_name}-worker"
  role             = aws_iam_role.worker.arn
  handler          = "bootstrap"
  runtime          = "provided.al2023"
  architectures    = ["x86_64"]
  # Must stay 1/6 of the queue's visibility timeout, per sqs.tf.
  timeout          = 30
  filename         = data.archive_file.worker.output_path
  source_code_hash = data.archive_file.worker.output_base64sha256

  environment {
    variables = {
      DYNAMODB_TABLE = aws_dynamodb_table.metadata.name
    }
  }
}

resource "aws_lambda_event_source_mapping" "worker" {
  event_source_arn = aws_sqs_queue.jobs.arn
  function_name    = aws_lambda_function.worker.arn
  # One job per invocation keeps failure handling simple: a batch >1 would
  # retry the whole batch when a single message fails.
  batch_size = 1
}
