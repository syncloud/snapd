package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
	"os"
)

const (
	Bucket    = "apps.syncloud.org"
	AwsKey    = "AWS_ACCESS_KEY_ID"
	AwsSecret = "AWS_SECRET_ACCESS_KEY"
)

func Upload(file io.Reader, name string) error {
	awsKeyValue, present := os.LookupEnv(AwsKey)
	if !present {
		return fmt.Errorf("%s env variable is not set", AwsKey)
	}
	awsSecretValue, present := os.LookupEnv(AwsSecret)
	if !present {
		return fmt.Errorf("%s env variable is not set", AwsSecret)
	}
	sess := session.Must(session.NewSession(
		&aws.Config{
			Credentials: credentials.NewStaticCredentials(
				awsKeyValue,
				awsSecretValue,
				"",
			),
			Region: aws.String("us-west-2"),
		},
	))

	uploader := s3manager.NewUploader(sess)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(Bucket),
		ACL:    aws.String("public-read"),
		Key:    aws.String(name),
		Body:   file,
	})
	return err
}
