resource "aws_cloudwatch_event_bus" "parcel" {
  name = "${var.project_name}-events"
}

resource "aws_cloudwatch_event_rule" "processing_completed" {
  name           = "${var.project_name}-processing-completed"
  event_bus_name = aws_cloudwatch_event_bus.parcel.name

  event_pattern = jsonencode({
    source      = ["parcel.worker"]
    detail-type = ["ProcessingCompleted"]
  })
}

resource "aws_cloudwatch_event_target" "sns" {
  rule           = aws_cloudwatch_event_rule.processing_completed.name
  event_bus_name = aws_cloudwatch_event_bus.parcel.name
  target_id      = "SNSProcessingNotifications"
  arn            = aws_sns_topic.processing.arn
}

resource "aws_sns_topic_policy" "eventbridge" {
  arn = aws_sns_topic.processing.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "events.amazonaws.com" }
      Action    = "sns:Publish"
      Resource  = aws_sns_topic.processing.arn
      Condition = {
        ArnLike = {
          "aws:SourceArn" = aws_cloudwatch_event_rule.processing_completed.arn
        }
      }
    }]
  })
}