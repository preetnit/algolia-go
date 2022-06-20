package aws

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type S3Object struct {
	Key  string `mapstructure:"key"`
	Size int    `mapstructure:"size"`
}

type S3Schema struct {
	Object S3Object `mapstructure:"object"`
}

type S3Record struct {
	S3 S3Schema `mapstructure:"s3"`
}

type S3Messages struct {
	Records []S3Record `mapstructure:"Records"`
}

func GetSQSQueueURL(sess *session.Session, queue string) (*sqs.GetQueueUrlOutput, error) {
	sqsClient := sqs.New(sess)

	result, err := sqsClient.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: &queue,
	})

	if err != nil {
		if err != nil {
			fmt.Printf("Got an error while trying to create queue: %v", err)
			return nil, err
		}
	}

	return result, nil
}

func GetMessages(sess *session.Session, queueUrl string, maxMessages int) *sqs.ReceiveMessageOutput {
	sqsClient := sqs.New(sess)

	msgResult, err := sqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:            &queueUrl,
		MaxNumberOfMessages: aws.Int64(int64(maxMessages)),
		WaitTimeSeconds:     aws.Int64(5),
	})

	if err != nil {
		fmt.Printf("Got an error while trying to retrieve message: %v", err)
		return nil
	} else if msgResult.Messages == nil {
		fmt.Printf("No new messages retrieved from Queue")
		return nil
	} else {
		fmt.Println("Message Body: " + *msgResult.Messages[0].Body)
		return msgResult
	}
}

func GetFileNameFromMessage(msgResult *sqs.ReceiveMessageOutput) string {
	var data S3Messages
	json.Unmarshal([]byte(*msgResult.Messages[0].Body), &data)
	fmt.Printf("FileName is %v\n", data.Records[0].S3.Object.Key)
	return data.Records[0].S3.Object.Key
}

func DeleteMessage(sess *session.Session, queueUrl string, message *sqs.ReceiveMessageOutput) {
	fmt.Printf("Deleting Message in Queue: %v\n with Message Handle: %v\n", queueUrl, &message.Messages[0].ReceiptHandle)
	sqsClient := sqs.New(sess)
	sqsClient.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      &queueUrl,
		ReceiptHandle: message.Messages[0].ReceiptHandle,
	})
	fmt.Println("Deleted Successfully")
}

func listQueues(sess *session.Session, queueName string) {
	sqsClient := sqs.New(sess)
	result, err := sqsClient.ListQueues(&sqs.ListQueuesInput{QueueNamePrefix: &queueName})
	if err != nil {
		fmt.Println(err)
	}

	for i, url := range result.QueueUrls {
		fmt.Printf("%v: %v\n", i, *url)
	}
}
