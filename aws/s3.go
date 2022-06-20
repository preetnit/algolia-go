package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/preetnit/algolia-go/algolia"
	"github.com/preetnit/algolia-go/config"
	"io"
)

func ListBucketObjects(sess *session.Session, bucketName string) {
	svc := s3.New(sess)
	input := &s3.ListObjectsInput{
		Bucket:  aws.String(bucketName),
		MaxKeys: aws.Int64(2),
	}

	result, err := svc.ListObjects(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				fmt.Println(s3.ErrCodeNoSuchBucket, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}

	fmt.Println(result)
}

func ReadS3File(sess *session.Session, cfg *config.Application, fileName string) (err error) {
	svc := s3.New(sess)

	params := &s3.SelectObjectContentInput{
		Bucket:         aws.String(cfg.S3BucketName),
		Key:            aws.String(fileName),
		ExpressionType: aws.String(s3.ExpressionTypeSql),
		Expression:     aws.String("SELECT r.objectID, r.hot_keywords, r.product_importance_score FROM S3Object r"),
		InputSerialization: &s3.InputSerialization{
			CSV: &s3.CSVInput{
				FileHeaderInfo: aws.String(s3.FileHeaderInfoUse),
			},
		},
		OutputSerialization: &s3.OutputSerialization{
			JSON: &s3.JSONOutput{},
		},
	}

	resp, err := svc.SelectObjectContent(params)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer resp.EventStream.Close()

	results, resultWriter := io.Pipe()
	go func() {
		defer resultWriter.Close()
		for event := range resp.EventStream.Events() {
			switch e := event.(type) {
			case *s3.RecordsEvent:
				resultWriter.Write(e.Payload)
			}
		}
	}()

	algolia.ProcessData(results, cfg)

	if err := resp.EventStream.Err(); err != nil {
		return fmt.Errorf("failed to read from SelectObjectContent EventStream, %v", err)
	}
	return nil
}
