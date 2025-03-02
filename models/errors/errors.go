package errors

import (
	"fmt"
)

type BaseError struct {
	MessageFmt string
	FmtItems   []interface{}
	ErrorCode  int
	Err        error
}

func (err BaseError) error() string {
	return fmt.Sprintf(err.MessageFmt, err.FmtItems...)
}

type NotImplementedError struct {
	BaseError
	name string
}

func (err NotImplementedError) Error() string {
	err.ErrorCode = 500
	err.MessageFmt = "Method '%s' is unimplemented"
	err.FmtItems = []interface{}{err.name}
	return err.error()
}

func RaiseNotImplErr(name string) error {
	return NotImplementedError{
		name: name,
	}
}

//
// Cloud Errors
//

type InitCloudError struct {
	BaseError
	conf map[string]interface{}
}

func (err InitCloudError) Error() string {
	err.ErrorCode = 500
	err.MessageFmt = "Unable to init Cloud Service => %s"
	err.FmtItems = []interface{}{err.Err}
	return err.error()
}

func RaiseInitCloudError(err error) error {
	ex := InitCloudError{}
	ex.Err = err
	return ex
}

type ListObjectsError struct {
	BaseError
	bucket string
	prefix string
}

func (err ListObjectsError) Error() string {
	err.ErrorCode = 500
	err.MessageFmt = "Unable to listObjects at 's3://%s/%s' => %s"
	err.FmtItems = []interface{}{err.bucket, err.prefix, err.Err}
	return err.error()
}

func RaiseListObjsErr(bucket, prefix string, err error) error {
	ex := ListObjectsError{
		bucket: bucket,
		prefix: prefix,
	}

	ex.Err = err
	return ex
}
