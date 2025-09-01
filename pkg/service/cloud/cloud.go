package cloud

import (
	"context"
	"fmt"
	"os"

	"github.com/KyloRilo/helios/pkg/model/errors"
)

type CloudConfig struct {
	Cloud      string
	Region     string
	RoleArn    string
	ExternalId string
	AccessKey  string
	SecretKey  string
}

func NewCloudConfig() CloudConfig {
	return CloudConfig{
		Cloud:      os.Getenv("CLOUD_PROVIDER"),
		Region:     os.Getenv("REGION"),
		RoleArn:    os.Getenv("ROLE_ARN"),
		ExternalId: os.Getenv("EXTERNAL_ID"),
		AccessKey:  os.Getenv("ACCESS_KEY"),
		SecretKey:  os.Getenv("SECRET_KEY"),
	}
}

type ICloudStoreService interface {
	Init(context.Context, CloudConfig) error
	PutObject(context.Context, string, string, string) error
	GetObject(context.Context, string, string) (string, error)
	ListObjects(context.Context, string, string) ([]string, error)
}

type CloudStorageService struct{}

func (cl CloudStorageService) Init(ctx context.Context, conf CloudConfig) error {
	return errors.RaiseNotImplErr("Init")
}

func (cl CloudStorageService) GetObject(ctx context.Context, bucket, key string) (string, error) {
	return "", errors.RaiseNotImplErr("GetObject")
}

func (cl CloudStorageService) PutObject(ctx context.Context, bucket string, key string, body string) error {
	return errors.RaiseNotImplErr("PutObject")
}

func (cl CloudStorageService) ListObjects(ctx context.Context, bucket, prefix string) ([]string, error) {
	return nil, errors.RaiseNotImplErr("ListObjects")
}

const (
	GCP = "gcp"
	AWS = "aws"
)

func InitCloudStorageService(conf CloudConfig) (ICloudStoreService, error) {
	var cloudService ICloudStoreService
	ctx := context.Background()

	switch conf.Cloud {
	case GCP:
		cloudService = gcpStorageService{}
	case AWS:
		cloudService = awsStorageService{
			roleArn:    conf.RoleArn,
			region:     conf.Region,
			externalId: conf.ExternalId,
		}
	default:
		return nil, nil
	}

	if err := cloudService.Init(ctx, conf); err != nil {
		return nil, fmt.Errorf("failed to init cloud service => %s", err)
	}

	return cloudService, nil
}
