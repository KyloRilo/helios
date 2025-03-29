package models

import (
	"context"

	"github.com/KyloRilo/helios/models/errors"

	"github.com/google/uuid"
)

type CloudConfig struct {
	Cloud      string
	Region     string
	RoleArn    string
	ExternalId string
	AccessKey  string
	PrivateKey string
}

type CloudServiceIntf interface {
	Init(context.Context, CloudConfig) error
	GetObject(_ context.Context, bucket, key string) (string, error)
	ListObjects(_ context.Context, bucket, prefix string) ([]string, error)
}

type CloudService struct {
	Uuid string
}

func (cl *CloudService) init(_ context.Context, _ CloudConfig) error {
	return errors.RaiseNotImplErr("init")
}

func (cl CloudService) Init(ctx context.Context, conf CloudConfig) error {
	cl.Uuid = uuid.New().String()
	return cl.init(ctx, conf)
}

func (cl *CloudService) getObject(_ context.Context, _, _ string) (string, error) {
	return "", errors.RaiseNotImplErr("getObject")
}

func (cl CloudService) GetObject(ctx context.Context, bucket, key string) (string, error) {
	return cl.getObject(ctx, bucket, key)
}

func (cl CloudService) listObjects(_ context.Context, _, _ string) ([]string, error) {
	return nil, errors.RaiseNotImplErr("listObjects")
}

func (cl CloudService) ListObjects(ctx context.Context, bucket, prefix string) ([]string, error) {
	return cl.listObjects(ctx, bucket, prefix)
}
