package redis

type FollowerReplication struct {
	Replication
	MasterHost string
	MasterPort string
}