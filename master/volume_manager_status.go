package master

import "time"

const MaxHeartbeatDuration time.Duration = time.Second * 10

type VolumeManagerStatus struct {
	AdminHost     string
	AdminPort     int
	PublicHost    string
	PublicPort    int

	DiskSize      uint64
	DiskUsed      uint64
	DiskFree      uint64

	LastHeartbeat time.Time `json:"-"`

	VStatusList   []*VolumeStatus
}

func (vms *VolumeManagerStatus)IsAlive() bool {
	return vms.LastHeartbeat.Add(MaxHeartbeatDuration).After(time.Now())
}
