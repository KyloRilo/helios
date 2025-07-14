package errors

import (
	"fmt"
)

type BaseError struct {
	ErrorCode int
	Err       error
}

func (err BaseError) error(errorFmt string, items ...interface{}) string {
	return fmt.Sprintf(errorFmt, items...)
}

type NotImplementedError struct {
	BaseError
	name string
}

func (err NotImplementedError) Error() string {
	return err.error(
		"Method '%s' is unimplemented",
		err.name,
	)
}

func RaiseNotImplErr(name string) error {
	return NotImplementedError{
		name: name,
		BaseError: BaseError{
			ErrorCode: 500,
		},
	}
}

//
// Cloud Errors
//

type InitCloudError struct {
	BaseError
}

func (err InitCloudError) Error() string {
	return err.error(
		"Unable to init Cloud Service => %s",
		err.Err,
	)
}

func RaiseInitCloudError(err error) error {
	return InitCloudError{
		BaseError: BaseError{
			Err:       err,
			ErrorCode: 500,
		},
	}
}

type ListObjectsError struct {
	BaseError
	bucket string
	prefix string
}

func (err ListObjectsError) Error() string {
	return err.error(
		"Unable to listObjects at 's3://%s/%s' => %s",
		err.bucket, err.prefix, err.Err,
	)
}

func RaiseListObjsErr(bucket, prefix string, err error) error {
	return ListObjectsError{
		bucket: bucket,
		prefix: prefix,
		BaseError: BaseError{
			Err:       err,
			ErrorCode: 500,
		},
	}
}

type GetObjectError struct {
	BaseError
	bucket string
	key    string
}

func (err GetObjectError) Error() string {
	return err.error(
		"Unable to perform GetObject => Bucket: '%s', Key: '%s', => %s",
		err.bucket, err.key, err.Err,
	)
}

func RaiseGetObjErr(bucket string, key string, err error) error {
	return GetObjectError{
		bucket: bucket,
		key:    key,
		BaseError: BaseError{
			ErrorCode: 500,
			Err:       err,
		},
	}
}

type PutObjectError struct {
}
