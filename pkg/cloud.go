package pkg

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/KyloRilo/helios/models"
	"github.com/KyloRilo/helios/models/errors"
	"github.com/KyloRilo/helios/proto"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	GCP = "gcp"
	AWS = "aws"
)

type awsProvider struct {
	models.CloudService
	conf       aws.Config
	s3Client   *s3.Client
	roleArn    string
	region     string
	externalId string
}

func (prv awsProvider) init(ctx context.Context, conf models.CloudConfig) error {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(conf.Region),
		config.WithAssumeRoleCredentialOptions(func(options *stscreds.AssumeRoleOptions) {
			options.RoleARN = conf.RoleArn
			options.RoleSessionName = fmt.Sprintf("storage-reader-%s", prv.Uuid)
			options.ExternalID = &conf.ExternalId
		}),
	)

	if err != nil {
		return errors.RaiseInitCloudError(err)
	}

	prv.conf = cfg
	prv.s3Client = s3.NewFromConfig(cfg)
	return nil
}

func (prv awsProvider) listObjects(ctx context.Context, bucket, prefix string) ([]string, error) {
	output, err := prv.s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})

	if err != nil {
		errors.RaiseListObjsErr(bucket, prefix, err)
	}

	fmt.Print(output)
	return nil, nil
}

type gcpProvider struct {
	models.CloudService
}

//
// gRPC impl
//

var cloudService models.CloudServiceIntf

const cloudServerPort int = 50502

type CloudServer struct {
	proto.UnimplementedCloudServer
}

func (s *CloudServer) ListObjects(ctx context.Context, in *proto.ListReq) (*proto.ListResp, error) {
	cloudService.ListObjects(ctx, in.Bucket, in.Prefix)
	return nil, nil
}

func InitCloudService(conf models.CloudConfig) {
	var err error
	ctx := context.Background()

	switch conf.Cloud {
	case GCP:
		cloudService = gcpProvider{}
	case AWS:
		cloudService = awsProvider{
			roleArn:    conf.RoleArn,
			region:     conf.Region,
			externalId: conf.ExternalId,
		}
	}

	if err := cloudService.Init(ctx, conf); err != nil {
		log.Fatal("failed to init cloud service => ", err)
	}

	lis, err := net.Listen("tcp", formatPort(cloudServerPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	proto.RegisterCloudServer(s, &CloudServer{})
	reflection.Register(s)

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
