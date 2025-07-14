package raft

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"github.com/hashicorp/raft"
	boltdb "github.com/hashicorp/raft-boltdb"
)

type ManagerState struct {
	state *map[string]interface{}
}

func (s *ManagerState) Persist(sink raft.SnapshotSink) error {
	var bytes []byte
	var err error
	if bytes, err = json.Marshal(s.state); err != nil {
		return fmt.Errorf("ManagerState.Persist() => Failed to convert state to json => %s", err)
	}

	if _, err = sink.Write(bytes); err != nil {
		sink.Cancel()
		return fmt.Errorf("ManagerState.Persist() => Failed to write state => %s", err)
	}

	return sink.Close()
}

func (s ManagerState) Restore(r io.ReadCloser) error {
	var bytes []byte
	var err error

	if bytes, err = io.ReadAll(r); err != nil {
		log.Panicf("ManagerState.Restore() => Unable to read bytes => %s", err)
	}

	if err = json.Unmarshal(bytes, s.state); err != nil {
		log.Panicf("ManagerState.Restore() => Unable to apply state => %s", err)
	}

	return nil
}

func (s ManagerState) Apply(rlog *raft.Log) interface{} {
	var err error
	if err = json.Unmarshal(rlog.Data, s.state); err != nil {
		log.Panicf("ManagerState.Restore() => Failed to apply state json => %s", err)
	}

	return nil
}

func (s ManagerState) Snapshot() (raft.FSMSnapshot, error) {
	return nil, nil
}

type Manager struct {
	mtx     sync.Mutex
	address raft.ServerAddress
	state   ManagerState
}

func (m *Manager) Transport() raft.Transport {
	return nil
}

func InitRaftManager() *Manager {
	raftDir := os.Getenv("RAFT_DATA_DIR")
	id := uuid.New()
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(id.String())

	baseDir := filepath.Join(raftDir, id.String())

	logDb, err := boltdb.NewBoltStore(filepath.Join(baseDir, "logs.dat"))
	if err != nil {
		log.Panicf(`boltdb.NewBoltStore(%q): %v`, filepath.Join(baseDir, "logs.dat"), err)
	}

	stableDb, err := boltdb.NewBoltStore(filepath.Join(baseDir, "stable.dat"))
	if err != nil {
		log.Panicf(`boltdb.NewBoltStore(%q): %v`, filepath.Join(baseDir, "stable.dat"), err)
	}

	fileSnap, err := raft.NewFileSnapshotStore(baseDir, 3, os.Stderr)
	if err != nil {
		log.Panicf(`raft.NewFileSnapshotStore(%q, ...): %v`, baseDir, err)
	}

	mngr := &Manager{}
	r, err := raft.NewRaft(config, mngr.state, logDb, stableDb, fileSnap, mngr.Transport())
	if err != nil {
		log.Panicf("raft.InitRaftManager() => Unable to create new raft => %s", err)
	}

	cfg := raft.Configuration{
		Servers: []raft.Server{{
			Suffrage: raft.Voter,
			ID:       raft.ServerID(""),
			Address:  raft.ServerAddress(""),
		}},
	}

	future := r.BootstrapCluster(cfg)
	if err := future.Error(); err != nil {
		log.Panicf("raft.InitRaftManager() => Unable to bootstrap raft cluster => %s", err)
	}

	return mngr
}
