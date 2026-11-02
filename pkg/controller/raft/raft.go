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

type RaftServiceState struct {
	mtx   sync.RWMutex
	state map[string]interface{}
}

func (s *RaftServiceState) Persist(sink raft.SnapshotSink) error {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	bytes, err := json.Marshal(s.state)
	if err != nil {
		sink.Cancel()
		return fmt.Errorf("RaftServiceState.Persist: marshal: %w", err)
	}

	if _, err = sink.Write(bytes); err != nil {
		sink.Cancel()
		return fmt.Errorf("RaftServiceState.Persist: write: %w", err)
	}

	return sink.Close()
}

func (s *RaftServiceState) Restore(r io.ReadCloser) error {
	defer r.Close()

	bytes, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("RaftServiceState.Restore: read: %w", err)
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	if err = json.Unmarshal(bytes, &s.state); err != nil {
		return fmt.Errorf("RaftServiceState.Restore: unmarshal: %w", err)
	}

	return nil
}

func (s *RaftServiceState) Apply(rlog *raft.Log) interface{} {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if err := json.Unmarshal(rlog.Data, &s.state); err != nil {
		log.Printf("RaftServiceState.Apply: unmarshal: %v", err)
		return err
	}

	return nil
}

func (s *RaftServiceState) Snapshot() (raft.FSMSnapshot, error) {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	clone := make(map[string]interface{}, len(s.state))
	for k, v := range s.state {
		clone[k] = v
	}

	return &RaftServiceState{state: clone}, nil
}

func (s *RaftServiceState) Release() {}

type RaftService struct {
	mtx   sync.Mutex
	conf  raft.Server
	raft  *raft.Raft
	state *RaftServiceState
}

func (m *RaftService) IsLeader() bool {
	return m.raft.State() == raft.Leader
}

func (m *RaftService) LeaderAddress() (raft.ServerAddress, error) {
	addr, _ := m.raft.LeaderWithID()
	if addr == "" {
		return "", fmt.Errorf("no leader elected")
	}
	return addr, nil
}

func (m *RaftService) AddVoter(id string, address string) error {
	future := m.raft.AddVoter(raft.ServerID(id), raft.ServerAddress(address), 0, 0)
	return future.Error()
}

func (m *RaftService) RemoveServer(id string) error {
	future := m.raft.RemoveServer(raft.ServerID(id), 0, 0)
	return future.Error()
}

func (m *RaftService) BootstrapCluster(servers []raft.Server) error {
	future := m.raft.BootstrapCluster(raft.Configuration{
		Servers: servers,
	})
	return future.Error()
}

func (m *RaftService) GetRaft() *raft.Raft {
	return m.raft
}

type RaftConfig struct {
	NodeID  string
	RaftDir string
	Bind    string
}

func InitRaftService(cfg RaftConfig) (*RaftService, error) {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(cfg.NodeID)

	if err := os.MkdirAll(cfg.RaftDir, 0700); err != nil {
		return nil, fmt.Errorf("raft: mkdir %s: %w", cfg.RaftDir, err)
	}

	logDb, err := boltdb.NewBoltStore(filepath.Join(cfg.RaftDir, "logs.dat"))
	if err != nil {
		return nil, fmt.Errorf("raft: log store: %w", err)
	}

	stableDb, err := boltdb.NewBoltStore(filepath.Join(cfg.RaftDir, "stable.dat"))
	if err != nil {
		return nil, fmt.Errorf("raft: stable store: %w", err)
	}

	fileSnap, err := raft.NewFileSnapshotStore(cfg.RaftDir, 3, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("raft: snapshot store: %w", err)
	}

	transport, err := raft.NewTCPTransport(cfg.Bind, nil, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("raft: tcp transport on %s: %w", cfg.Bind, err)
	}

	state := &RaftServiceState{
		state: make(map[string]interface{}),
	}

	svc := &RaftService{
		state: state,
		conf: raft.Server{
			Suffrage: raft.Voter,
			ID:       config.LocalID,
			Address:  raft.ServerAddress(cfg.Bind),
		},
	}

	raftRef, err := raft.NewRaft(config, state, logDb, stableDb, fileSnap, transport)
	if err != nil {
		return nil, fmt.Errorf("raft: init: %w", err)
	}

	svc.raft = raftRef
	return svc, nil
}
