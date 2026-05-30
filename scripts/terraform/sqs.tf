resource "aws_sqs_queue" "events-sqs-queue" {
  name = "events-sqs-queue"
}

resource "aws_sns_topic_subscription" "events-sns-sqs-subscription" {
  topic_arn = aws_sns_topic.events-sns-topic.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.events-sqs-queue.arn 
}