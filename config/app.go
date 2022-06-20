package config

import (
	"github.com/spf13/viper"
	"log"
)

type Application struct {
	AlgoliaAppId        string `mapstructure:"ALGOLIA_APP_ID"`
	AlgoliaAPIKey       string `mapstructure:"ALGOLIA_API_KEY"`
	AlgoliaIndexName    string `mapstructure:"ALGOLIA_INDEX_NAME"`
	AlgoliaOpsBatchSize int    `mapstructure:"ALGOLIA_OPS_BATCH_SIZE"`
	SQSReadInterval     int    `mapstructure:"SQS_READ_INTERVAL_MINUTES"`
	SQSQueueName        string `mapstructure:"SQS_QUEUE_NAME"`
	S3BucketName        string `mapstructure:"S3_BUCKET_NAME"`
}

var App *Application

func Load() {
	viper.SetConfigFile(".env")
	viper.ReadInConfig()

	App = new(Application)
	if err := viper.Unmarshal(&App); err != nil {
		log.Fatal(err)
	}
}
