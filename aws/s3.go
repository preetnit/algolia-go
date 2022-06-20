package aws

import (
	"encoding/json"
	"fmt"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/preetnit/algolia-go/algolia"
	"github.com/preetnit/algolia-go/config"
	"io"
	"sync"
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
		Expression:     aws.String("SELECT *  FROM S3Object"),
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

	var wg sync.WaitGroup
	resReader := json.NewDecoder(results)
	var operations []search.BatchOperationIndexed
	var record algolia.Record
	for {
		err := resReader.Decode(&record)
		if err == io.EOF {
			fmt.Println("EOF")
			wg.Add(1)
			go algolia.UpdateIndex(cfg, operations, &wg)
			break
		}

		fmt.Printf("Record %v\n", record)
		operations = append(operations, search.BatchOperationIndexed{
			IndexName: cfg.AlgoliaIndexName,
			BatchOperation: search.BatchOperation{
				Action: search.PartialUpdateObjectNoCreate,
				Body:   record,
			},
		})

		if len(operations) == cfg.AlgoliaOpsBatchSize {
			wg.Add(1)
			go algolia.UpdateIndex(cfg, operations, &wg)
			operations = nil
		}
	}

	if err := resp.EventStream.Err(); err != nil {
		return fmt.Errorf("failed to read from SelectObjectContent EventStream, %v", err)
	}
	wg.Wait()
	return nil
}
