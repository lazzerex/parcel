resource "aws_sns_topic" "processing" {
  name = "${var.project_name}-processing-notifications"
}