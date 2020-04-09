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
		Name: "sqs_messages_visible",
		Help: "Type: Gauge, The number of available messages in queue(s). Use the name of the queue as the label to get the messages of a specific queue e.g `sqs_messages_visible{queue_name=\"<QUEUE NAME>\"}`.",
	}, []string{"queue_name"})
	delayedMessageGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sqs_messages_delayed",
		Help: "Type: Gauge, The number of messages waiting to be added into queue(s). Use the name of the queue as the label to get the messages of a specific queue e.g `sqs_messages_delayed{queue_name=\"<QUEUE NAME>\"}`.",
	}, []string{"queue_name"})
	invisibleMessageGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sqs_messages_invisible",
		Help: "Type: Gauge, The number of messages in flight in queue(s). Use the name of the queue as the label to get the messages of a specific queue e.g `sqs_messages_invisible{queue_name=\"<QUEUE NAME>\"}`.",
	}, []string{"queue_name"})
)

func init() {
	prometheus.MustRegister(visibleMessageGauge)
	prometheus.MustRegister(delayedMessageGauge)
	prometheus.MustRegister(invisibleMessageGauge)
}

// MonitorSQS Retrieves the attributes of all allowed queues from SQS and appends the metrics
func MonitorSQS() error {
	queues, _ , err := getQueues()
	if err != nil {
		return fmt.Errorf("Error occurred while retrieve queues info from SQS: %v", err)
	}

	for queue, attr := range queues {
		msgAvailable, msgError := strconv.ParseFloat(*attr.Attributes["ApproximateNumberOfMessages"], 64)
		msgDelayed, delayError := strconv.ParseFloat(*attr.Attributes["ApproximateNumberOfMessagesDelayed"], 64)
		msgNotVisible, invisibleError := strconv.ParseFloat(*attr.Attributes["ApproximateNumberOfMessagesNotVisible"], 64)
		
		if msgError != nil {
			return fmt.Errorf("Error in converting ApproximateNumberOfMessages: %v", msgError)
		}
		visibleMessageGauge.WithLabelValues(queue).Add(msgAvailable)

		if delayError != nil {
			return fmt.Errorf("Error in converting ApproximateNumberOfMessagesDelayed: %v", delayError)
		}
		delayedMessageGauge.WithLabelValues(queue).Add(msgDelayed)

		if invisibleError != nil {
			return fmt.Errorf("Error in converting ApproximateNumberOfMessagesNotVisible: %v", invisibleError)
		}
		invisibleMessageGauge.WithLabelValues(queue).Add(msgNotVisible)
	}
	return nil
}

func getQueueName(url string) (queueName string) {
	queue := strings.Split(url, "/")
	queueName = queue[len(queue)-1]
	return
}

func getQueues() (queues map[string]*sqs.GetQueueAttributesOutput, tags map[string]*sqs.ListQueueTagsOutput, err error) {
	sess := session.Must(session.NewSession())
	client := sqs.New(sess)
	result, err := client.ListQueues(nil)
	if err != nil {
		return nil, nil, err
	}
	fmt.Println(result)
	if result.QueueUrls == nil {
		err = fmt.Errorf("SQS did not return any QueueUrls")
		return nil, nil, err
	}

	queues = make(map[string]*sqs.GetQueueAttributesOutput)
	tags = make(map[string]*sqs.ListQueueTagsOutput)

	for _, urls := range result.QueueUrls {
		params := &sqs.GetQueueAttributesInput{
			QueueUrl: aws.String(*urls),
			AttributeNames: []*string{
				aws.String("ApproximateNumberOfMessages"),
				aws.String("ApproximateNumberOfMessagesDelayed"),
				aws.String("ApproximateNumberOfMessagesNotVisible"),
			},
		}

		tagsParams := &sqs.ListQueueTagsInput{
			QueueUrl: aws.String(*urls),
		}

		resp, err := client.GetQueueAttributes(params)
		if err != nil {
			return nil, nil, err
		}
		tagsResp, err := client.ListQueueTags(tagsParams)
		if err != nil {
			return nil, nil, err
		}
		queueName := getQueueName(*urls)
		queues[queueName] = resp
		tags[queueName] = tagsResp
	}
	return queues,tags, nil
}
