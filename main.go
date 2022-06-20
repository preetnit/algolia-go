package main

import (
	"fmt"
	"github.com/preetnit/algolia-go/aws"
	"github.com/preetnit/algolia-go/config"
	"time"
)

func main() {
	config.Load()
	sess := aws.InitSession()
	urlRes, err := aws.GetSQSQueueURL(sess, config.App.SQSQueueName)
	if err != nil {
		return
	}
	maxMessages := 1
	for {
		fmt.Printf("\n################\n")
		fmt.Printf("Polling SQS: %v\n", config.App.SQSQueueName)

		message := aws.GetMessages(sess, *urlRes.QueueUrl, maxMessages)
		if message != nil {
			fileName := aws.GetFileNameFromMessage(message)
			err := aws.ReadS3File(sess, config.App, fileName)

			if err == nil {
				aws.DeleteMessage(sess, *urlRes.QueueUrl, message)
			}
		}

		fmt.Println("\n+++++++++++++++++")
		fmt.Printf("Sleeping for %v Minutes\n", config.App.SQSReadInterval)
		time.Sleep(time.Minute * time.Duration(config.App.SQSReadInterval))
		fmt.Println("\n+++++++++++++++++")
	}
}
