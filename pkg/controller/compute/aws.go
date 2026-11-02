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

func (a AwsCtrl) createNetwork(_ context.Context, _ string) error { return nil }
func (a AwsCtrl) removeNetwork(_ context.Context, _ string) error { return nil }

func newAwsCtrl(ctx context.Context, creds AwsCreds) CtrlShim {
	return AwsCtrl{}
}
