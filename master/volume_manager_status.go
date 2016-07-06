package master

import (
	"time"
	"github.com/030io/whalefs/volume/api"
)

const MaxHeartbeatDuration time.Duration = time.Second * 10

type VolumeManagerStatus struct {
	AdminHost     string
	AdminPort     int
	PublicHost    string
	PublicPort    int

	Machine       string
	DataCenter    string

	DiskSize      uint64
	DiskUsed      uint64
	DiskFree      uint64

	LastHeartbeat time.Time `json:"-"`

	VStatusList   []*VolumeStatus
}

func (vms *VolumeManagerStatus)IsAlive() bool {
	return vms.LastHeartbeat.Add(MaxHeartbeatDuration).After(time.Now())
}

//TODO: 判断vm是否有足够的空间
func (vms *VolumeManagerStatus)canCreateVolume() bool {
	return true
}

func (vms *VolumeManagerStatus)createVolume(vid uint64) error {
	err := api.CreateVolume(vms.AdminHost, vms.AdminPort, vid)
	if err != nil {
		return err
	}

	vms.VStatusList = append(vms.VStatusList, &VolumeStatus{Id: vid, vmStatus: vms})
	return nil
}
