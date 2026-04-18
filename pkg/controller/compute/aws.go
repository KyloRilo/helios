package compute

import "context"

type AwsCreds struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

type AwsCtrl struct {
	ComputeController
}

func newAwsCtrl(ctx context.Context, creds AwsCreds) ComputeController {
	return AwsCtrl{}
}
