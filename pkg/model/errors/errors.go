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

type InvalidNodeStatus struct {
	NodeName   string
	NodeStatus string
	Expected   []string
	Action     string
}

func (err InvalidNodeStatus) Error() string {
	return fmt.Sprintf("'%s' Found with unexpected status '%s' while running action: %s => Expected: %v", err.NodeName, err.NodeStatus, err.Action, err.Expected)
}
