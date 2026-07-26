variable "account-id" {
  type = string
  default = "000000000000"
}

resource "aws_sns_topic" "events-sns-topic" {
  name = "events-sns-topic"
}

output "aws_sns_topic_arn" {
  value = aws_sns_topic.events-sns-topic.arn
}

output "aws_sns_topic_name" {
  value = aws_sns_topic.events-sns-topic.name
}
