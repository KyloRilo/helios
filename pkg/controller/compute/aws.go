package compute

import "context"

type AwsCreds struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

type AwsCtrl struct {
	CompImpl
}

func newAwsCtrl(ctx context.Context, creds AwsCreds) CtrlShim {
	return AwsCtrl{}
}
