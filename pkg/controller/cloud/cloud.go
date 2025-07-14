package cloud

import (
	"context"
	"fmt"

	"github.com/KyloRilo/helios/pkg/model/errors"
)

type CloudConfig struct {
	Cloud      string
	Region     string
	RoleArn    string
	ExternalId string
	AccessKey  string
	PrivateKey string
}

type ICloudStoreCtrl interface {
	Init(context.Context, CloudConfig) error
	PutObject(context.Context, string, string, string) error
	GetObject(context.Context, string, string) (string, error)
	ListObjects(context.Context, string, string) ([]string, error)
}

type CloudStorageCtrl struct{}

func (cl CloudStorageCtrl) Init(ctx context.Context, conf CloudConfig) error {
	return errors.RaiseNotImplErr("Init")
}

func (cl CloudStorageCtrl) GetObject(ctx context.Context, bucket, key string) (string, error) {
	return "", errors.RaiseNotImplErr("GetObject")
}

func (cl CloudStorageCtrl) PutObject(ctx context.Context, bucket string, key string, body string) error {
	return errors.RaiseNotImplErr("PutObject")
}

func (cl CloudStorageCtrl) ListObjects(ctx context.Context, bucket, prefix string) ([]string, error) {
	return nil, errors.RaiseNotImplErr("ListObjects")
}

const (
	GCP = "gcp"
	AWS = "aws"
)

func InitCloudStorageCtrl(conf CloudConfig) (ICloudStoreCtrl, error) {
	var cloudCtrl ICloudStoreCtrl
	ctx := context.Background()

	switch conf.Cloud {
	case GCP:
		cloudCtrl = gcpCtrl{}
	case AWS:
		cloudCtrl = awsStorageCtrl{
			roleArn:    conf.RoleArn,
			region:     conf.Region,
			externalId: conf.ExternalId,
		}
	default:
		return nil, fmt.Errorf("Unable to match cloud storage controller")
	}

	if err := cloudCtrl.Init(ctx, conf); err != nil {
		return nil, fmt.Errorf("failed to init cloud service => %s", err)
	}

	return cloudCtrl, nil
}
