package leader

import (
	"path/filepath"
	"testing"
)

func TestLeaderInit(t *testing.T) {
	path, err := filepath.Abs("../../../bin/helios/local.cluster.hcl")
	if err != nil {
		panic(err)
	}

	_ = NewLeader(t.Context(), path)
}
