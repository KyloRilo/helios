package unit

import (
	"path/filepath"
	"testing"

	"github.com/KyloRilo/helios/pkg/service/leader"
)

func TestLeaderInit(t *testing.T) {
	path, err := filepath.Abs("../config/helios.hcl")
	if err != nil {
		panic(err)
	}

	_ = leader.NewLeader(t.Context(), path)
}
