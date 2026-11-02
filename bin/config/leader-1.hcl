name      = "helios-leader-1"
host      = "helios-leader-1"
port      = 6331
node_id   = "1"
raft_dir  = "/raft/1"
raft_bind = "helios-leader-1:6330"
bootstrap = true

raft_peers = {
  "2" = "helios-leader-2:6330"
  "3" = "helios-leader-3:6330"
}
