package collector

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	visibleMessageGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sqs_approximatenumberofmessages",
		Help: "The approximate number of visible messages in a queue.",
	}, []string{"queue"})
	invisibleMessageGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sqs_approximatenumberofmessagesnotvisible",
		Help: "The approximate number of messages that have not timed-out and aren't deleted.",
	}, []string{"queue"})
)

func init() {
	prometheus.MustRegister(visibleMessageGauge)
	prometheus.MustRegister(invisibleMessageGauge)
}

// MonitorSQS Retrieves the attributes of all allowed queues from SQS and appends the metrics
func MonitorSQS(sqsNamePrefix, sqsEndpoint string) error {
	queues, err := getQueues(sqsNamePrefix, sqsEndpoint)
	if err != nil {
		return fmt.Errorf("[MONITORING ERROR]: Error occurred while retrieve queues info from SQS: %v", err)
	}

	for queue, attr := range queues {
		msgAvailable, msgError := strconv.ParseFloat(*attr.Attributes["ApproximateNumberOfMessages"], 64)
		msgNotVisible, invisibleError := strconv.ParseFloat(*attr.Attributes["ApproximateNumberOfMessagesNotVisible"], 64)

		if msgError != nil {
			return fmt.Errorf("Error in converting ApproximateNumberOfMessages: %v", msgError)
		}
		visibleMessageGauge.WithLabelValues(queue).Set(msgAvailable)

		if invisibleError != nil {
			return fmt.Errorf("Error in converting ApproximateNumberOfMessagesNotVisible: %v", invisibleError)
		}
		invisibleMessageGauge.WithLabelValues(queue).Set(msgNotVisible)
	}
	return nil
}

func getQueueName(url string) (queueName string) {
	queue := strings.Split(url, "/")
	queueName = queue[len(queue)-1]
	return
}

func getQueues(sqsNamePrefix, sqsEndpoint string) (queues map[string]*sqs.GetQueueAttributesOutput, err error) {
	sess := session.Must(session.NewSession())
	client := sqs.New(sess)
	awsCfg := aws.NewConfig()
	if sqsEndpoint != "" {
		awsCfg = awsCfg.WithEndpoint(sqsEndpoint)
	}
	client = sqs.New(sess, awsCfg)

	result, err := client.ListQueues(&sqs.ListQueuesInput{QueueNamePrefix: &sqsNamePrefix})
	if err != nil {
		return nil, err
	}
	if result.QueueUrls == nil {
		err = fmt.Errorf("SQS did not return any QueueUrls")
		return nil, err
	}

	queues = make(map[string]*sqs.GetQueueAttributesOutput)
	for _, urls := range result.QueueUrls {
		params := &sqs.GetQueueAttributesInput{
			QueueUrl: aws.String(*urls),
			AttributeNames: []*string{
				aws.String("ApproximateNumberOfMessages"),
				aws.String("ApproximateNumberOfMessagesNotVisible"),
			},
		}

		resp, err := client.GetQueueAttributes(params)
		if err != nil {
			return nil, err
		}
		queueName := getQueueName(*urls)
		queues[queueName] = resp
	}
	return queues, nil
}
