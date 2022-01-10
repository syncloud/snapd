package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
	"os"
	"strings"
)

const (
	AwsKey    = "AWS_ACCESS_KEY_ID"
	AwsSecret = "AWS_SECRET_ACCESS_KEY"
)

type S3 struct {
	bucket string
}

func NewS3(bucket string) *S3 {
	return &S3{bucket: bucket}
}

func (s *S3) UploadFile(from string, to string) error {
	reader, err := os.Open(from)
	Check(err)
	defer reader.Close()
	return s.upload(reader, to)
}

func (s *S3) UploadContent(content string, to string) error {
	return s.upload(strings.NewReader(content), to)
}

func (s *S3) upload(file io.Reader, name string) error {
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
		Bucket: aws.String(s.bucket),
		ACL:    aws.String("public-read"),
		Key:    aws.String(name),
		Body:   file,
	})
	return err
}
