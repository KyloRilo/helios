package cloud

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type awsStorageService struct {
	CloudStorageService
	s3Client   *s3.Client
	roleArn    string
	region     string
	externalId string
}

func (ctrl awsStorageService) Init(ctx context.Context, conf CloudConfig) error {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(ctrl.region),
		config.WithAssumeRoleCredentialOptions(func(options *stscreds.AssumeRoleOptions) {
			options.RoleARN = ctrl.roleArn
			options.RoleSessionName = fmt.Sprintf("storage-reader-%s", ctrl.externalId)
			options.ExternalID = &ctrl.externalId
		}),
	)

	if err != nil {
		return fmt.Errorf("awsStorageCtrl.Init() Unable to init S3 client => %s", err)
	}

	ctrl.s3Client = s3.NewFromConfig(cfg)
	return nil
}

func (ctrl awsStorageService) ListObjects(ctx context.Context, bucket, prefix string) ([]string, error) {
	keys := []string{}
	output, err := ctrl.s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})

	if err != nil {
		return nil, fmt.Errorf("awsStorageCtrl.ListObjects() Unable to list => %s", err)
	}

	for _, val := range output.Contents {
		keys = append(keys, *val.Key)
	}

	log.Printf("awsStorageCtrl.ListObjects() => Successfully ListObject on s3://%s/%s", bucket, prefix)
	return keys, nil
}

func (ctrl awsStorageService) GetObject(ctx context.Context, bucket, key string) (string, error) {
	output, err := ctrl.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return "", fmt.Errorf("awsStorageCtrl.GetObject() Unable to get object => %s", err)
	}
	defer output.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(output.Body)
	if err != nil {
		return "", fmt.Errorf("awsStorageCtrl.GetObject() Unable to read object => %s", err)
	}

	log.Printf("awsStorageCtrl.GetObject() => Successfully Got s3://%s/%s", bucket, key)
	return buf.String(), nil
}

func (ctrl awsStorageService) PutObject(ctx context.Context, bucket string, key string, body string) error {
	_, err := ctrl.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader([]byte(body)),
	})

	if err != nil {
		return fmt.Errorf("awsStorageCtrl.PutObject() Unable to put object => %s", err)
	}

	log.Printf("awsStorageCtrl.PutObject() => Successfully Uploaded s3://%s/%s", bucket, key)
	return nil
}
