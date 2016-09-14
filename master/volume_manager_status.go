package master

import (
	"time"
	"github.com/030io/whalefs/manager/api"
)

var MaxHeartbeatDuration time.Duration = time.Second * 10

type VolumeManagerStatus struct {
	AdminHost       string
	AdminPort       int
	PublicHost      string
	PublicPort      int

	Machine         string
	DataCenter      string

	DiskSize        uint64
	DiskUsed        uint64
	DiskFree        uint64
	MaxDiskUsed     uint64
	CanCreateVolume bool

	VolumeMaxSize   uint64

	LastHeartbeat   time.Time `json:"-"`

	VStatusList     []*VolumeStatus
}

func (vms *VolumeManagerStatus)IsAlive() bool {
	return vms.LastHeartbeat.Add(MaxHeartbeatDuration).After(time.Now())
}

func (vms *VolumeManagerStatus)createVolume(vid uint64) error {
	err := api.CreateVolume(vms.AdminHost, vms.AdminPort, vid)
	if err != nil {
		return err
	}

	vms.VStatusList = append(
		vms.VStatusList,
		&VolumeStatus{Id: vid, vmStatus: vms, Writable: true, MaxFreeSpace: 512 * 1 << 30},
	)
	return nil
}
