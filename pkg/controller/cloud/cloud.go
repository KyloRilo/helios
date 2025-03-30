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

type ICloudCtrl interface {
	Init(context.Context, CloudConfig) error
	GetObject(_ context.Context, bucket, key string) (string, error)
	ListObjects(_ context.Context, bucket, prefix string) ([]string, error)
}

type CloudController struct{}

func (cl CloudController) Init(ctx context.Context, conf CloudConfig) error {
	return errors.RaiseNotImplErr("init")
}

func (cl CloudController) GetObject(ctx context.Context, bucket, key string) (string, error) {
	return "", errors.RaiseNotImplErr("getObject")
}

func (cl CloudController) ListObjects(ctx context.Context, bucket, prefix string) ([]string, error) {
	return nil, errors.RaiseNotImplErr("listObjects")
}

const (
	GCP = "gcp"
	AWS = "aws"
)

func InitCloudController(conf CloudConfig) (ICloudCtrl, error) {
	var cloudCtrl ICloudCtrl
	ctx := context.Background()

	switch conf.Cloud {
	case GCP:
		cloudCtrl = gcpCtrl{}
	case AWS:
		cloudCtrl = awsCtrl{
			roleArn:    conf.RoleArn,
			region:     conf.Region,
			externalId: conf.ExternalId,
		}
	default:
		return nil, fmt.Errorf("Unable to match cloud controller")
	}

	if err := cloudCtrl.Init(ctx, conf); err != nil {
		return nil, fmt.Errorf("failed to init cloud service => %s", err)
	}

	return cloudCtrl, nil
}
