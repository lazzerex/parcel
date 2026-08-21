resource "aws_sqs_queue" "jobs_dlq" {
  name                      = local.dlq_name
  message_retention_seconds = 1209600
}

resource "aws_sqs_queue" "jobs" {
  name = local.queue_name

  # Must stay at least six times the worker's function timeout.
  visibility_timeout_seconds = 180

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.jobs_dlq.arn
    maxReceiveCount     = 3
  })
}
