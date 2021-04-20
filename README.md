# AWS SQS Prometheus Exporter

A Prometheus metrics exporter for AWS SQS queues

> **A few words of Thanks:** Most of the code in this repo is borrowed from [ashiddo11/sqs-exporter](https://github.com/ashiddo11/sqs-exporter) with bundle of thanks and love :pray: :heart:.

We didn't submit this as a pull request to the original repository, since:
1.  we have added Prometheus client. Whereas, some users of the original repository may not be using Prometheus at all.
2.  the owner of the original repository seems to be no longer actively maintaining the repo, and we needed the change as soon as possible.

## Metrics

| Metric  | Labels | Description |
| ------  | ------ | ----------- |
| sqs\_messages\_visible | Queue Name | Number of messages available |
| sqs\_messages\_delayed | Queue Name | Number of messages delayed |
| sqs\_messages\_not\_visible | Queue Name | Number of messages in flight |

For more information see the [AWS SQS Documentation](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-message-attributes.html)

## Configuration

Credentials to AWS are provided in the following order:

- Environment variables (AWS\_ACCESS\_KEY\_ID and AWS\_SECRET\_ACCESS\_KEY)
- Shared credentials file (~/.aws/credentials)
- IAM role for Amazon EC2

For more information see the [AWS SDK Documentation](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html)

### AWS IAM permissions

The app needs sqs list and read access to the sqs policies

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": [
                "sqs:ListQueues",
                "sqs:GetQueueUrl",
                "sqs:ListDeadLetterSourceQueues",
                "sqs:GetQueueAttributes"
            ],
            "Resource": "*"
        }
    ]
}
```

## Environment Variables

| Variable      | Default Value | Description                                                  |
|---------------|:---------|:-------------------------------------------------------------|
| PORT          | 9434     | The port for metrics server                                  |
| INTERVAL      | 1        | The interval in minutes to get the status of SQS queues      |
| ENDPOINT      | metrics  | The metrics endpoint                                         |
| KEEP_RUNNING  | true     | The flag to terminate the service in case of monitoring error |
| AWS_SQS_ENDPOINT | | Optional sqs endpoint (for mocked sqs service)  |
| SQS_QUEUE_NAME_PREFIX | | Optional prefix queue name to filter |
| AWS_REGION | |  |


## Running

```docker run -e INTERVAL=5 -e KEEP_RUNNING=false -d -p 9434:9434 jesusfcr/sqs-prometheus-exporter```

You can provide the AWS credentials as environment variables depending upon your security rules configured in AWS;

```docker run -d -p 9384:9384 -e AWS_ACCESS_KEY_ID=<access_key> -e AWS_SECRET_ACCESS_KEY=<secret_key> -e AWS_REGION=<region>  jesusfcr/sqs-prometheus-exporter```
