package storage

// import (
// 	"context"
// 	"fmt"

// 	"github.com/KyloRilo/helios/pkg/model/errors"
// )

// type StorageService interface {
// 	Init(context.Context) error
// 	PutObject(context.Context, string, string, string) error
// 	GetObject(context.Context, string, string) (string, error)
// 	ListObjects(context.Context, string, string) ([]string, error)
// }

// type BaseStorageService struct{}

// func (cl BaseStorageService) Init(ctx context.Context) error {
// 	return errors.RaiseNotImplErr("Init")
// }

// func (cl BaseStorageService) GetObject(ctx context.Context, bucket, key string) (string, error) {
// 	return "", errors.RaiseNotImplErr("GetObject")
// }

// func (cl BaseStorageService) PutObject(ctx context.Context, bucket string, key string, body string) error {
// 	return errors.RaiseNotImplErr("PutObject")
// }

// func (cl BaseStorageService) ListObjects(ctx context.Context, bucket, prefix string) ([]string, error) {
// 	return nil, errors.RaiseNotImplErr("ListObjects")
// }

// const (
// 	GCP = "gcp"
// 	AWS = "aws"
// )

// func InitStorageController() (StorageService, error) {
// 	svc := BaseStorageService{}
// 	ctx := context.Background()

// 	if err := svc.Init(ctx); err != nil {
// 		return nil, fmt.Errorf("failed to init cloud service => %s", err)
// 	}

// 	return svc, nil
// }
