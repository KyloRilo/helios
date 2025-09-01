package raft

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

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
	mtx   sync.Mutex
	conf  raft.Server
	raft  *raft.Raft
	state ManagerState
}

func (m *Manager) IsLeader() bool {
	return m.raft.State() == raft.Leader
}

func (m *Manager) AddVoter() {
	// m.raft.AddVoter(m.conf.ID, m.conf.Address, 0, 0)
	m.raft.AddVoter("2", "helios-2:6330", 0, 0)
	m.raft.AddVoter("3", "helios-3:6330", 0, 0)
}

func (m *Manager) BootstrapCluster() {
	future := m.raft.BootstrapCluster(raft.Configuration{
		Servers: []raft.Server{m.conf},
	})
	if err := future.Error(); err != nil {
		log.Panicf("raft.InitRaftManager() => Unable to bootstrap raft cluster => %s", err)
	}
}

func InitRaftManager(nodeId string, raftDir string, host string) *Manager {
	fmt.Print("Raft Host: ", host)
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(nodeId)

	logDb, err := boltdb.NewBoltStore(filepath.Join(raftDir, "logs.dat"))
	if err != nil {
		log.Panicf(`boltdb.NewBoltStore(%q): %v`, filepath.Join(raftDir, "logs.dat"), err)
	}

	stableDb, err := boltdb.NewBoltStore(filepath.Join(raftDir, "stable.dat"))
	if err != nil {
		log.Panicf(`boltdb.NewBoltStore(%q): %v`, filepath.Join(raftDir, "stable.dat"), err)
	}

	fileSnap, err := raft.NewFileSnapshotStore(raftDir, 3, os.Stderr)
	if err != nil {
		log.Panicf(`raft.NewFileSnapshotStore(%q, ...): %v`, raftDir, err)
	}

	transport, err := raft.NewTCPTransport(host, nil, 3, 10*time.Second, os.Stderr)
	if err != nil {

	}

	mngr := &Manager{
		state: ManagerState{},
		conf: raft.Server{
			Suffrage: raft.Voter,
			ID:       config.LocalID,
			Address:  raft.ServerAddress(host),
		},
	}

	raftRef, err := raft.NewRaft(
		config,
		mngr.state,
		logDb,
		stableDb,
		fileSnap,
		transport,
	)

	if err != nil {
		log.Panicf("raft.InitRaftManager() => Unable to create new raft => %s", err)
	}

	mngr.raft = raftRef
	return mngr
}
