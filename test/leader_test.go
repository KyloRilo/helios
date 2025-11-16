package test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/KyloRilo/helios/pkg/service/leader"
)

func TestLeaderInit(t *testing.T) {
	ctx := context.Background()
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	path := filepath.Join(cwd, "config", "helios.hcl")
	fmt.Println(path)
	_ = leader.NewLeader(ctx, path)
}
