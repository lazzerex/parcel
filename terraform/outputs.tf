output "bucket_name" {
  value = aws_s3_bucket.files.bucket
}

output "table_name" {
  value = aws_dynamodb_table.metadata.name
}

output "queue_url" {
  value = aws_sqs_queue.jobs.url
}

output "queue_arn" {
  value = aws_sqs_queue.jobs.arn
}

output "dlq_url" {
  value = aws_sqs_queue.jobs_dlq.url
}

output "api_role_arn" {
  value = aws_iam_role.api.arn
}

output "worker_role_arn" {
  value = aws_iam_role.worker.arn
}
