package cloud

import (
	"context"
	"fmt"

	"github.com/KyloRilo/helios/pkg/model/errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type awsCtrl struct {
	CloudController
	conf       aws.Config
	s3Client   *s3.Client
	roleArn    string
	region     string
	externalId string
}

func (ctrl awsCtrl) Init(ctx context.Context, conf CloudConfig) error {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(conf.Region),
		config.WithAssumeRoleCredentialOptions(func(options *stscreds.AssumeRoleOptions) {
			options.RoleARN = conf.RoleArn
			options.RoleSessionName = fmt.Sprintf("storage-reader-%s", ctrl.externalId)
			options.ExternalID = &ctrl.externalId
		}),
	)

	if err != nil {
		return errors.RaiseInitCloudError(err)
	}

	ctrl.conf = cfg
	ctrl.s3Client = s3.NewFromConfig(cfg)
	return nil
}

func (ctrl awsCtrl) ListObjects(ctx context.Context, bucket, prefix string) ([]string, error) {
	output, err := ctrl.s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})

	if err != nil {
		errors.RaiseListObjsErr(bucket, prefix, err)
	}

	fmt.Print(output)
	return nil, nil
}
