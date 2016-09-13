package manager

import (
	"net/http"
	"github.com/030io/whalefs/manager/volume"
	"os"
	"io/ioutil"
	"strings"
	"strconv"
	"fmt"
	"time"
	"github.com/030io/whalefs/master/api"
	"github.com/030io/whalefs/master"
	"github.com/030io/whalefs/utils/disk"
)

var (
	MaxDiskUsedPercent uint = 99
	HeartbeatDuration time.Duration = time.Second * 5
	ReadOnly bool = false
)

type VolumeManager struct {
	DataDir      string
	Volumes      map[uint64]*volume.Volume

	AdminPort    int
	AdminHost    string
	PublicPort   int
	PublicHost   string
	AdminServer  *http.ServeMux
	PublicServer *http.ServeMux

	Machine      string
	DataCenter   string

	MasterHost   string
	MasterPort   int
}

func NewVolumeManager(dir string) (*VolumeManager, error) {
	f, err := os.OpenFile(dir, os.O_RDWR, 0)
	if os.IsNotExist(err) {
		panic(err)
	} else if os.IsPermission(err) {
		ReadOnly = true
	}
	f.Close()

	vm := new(VolumeManager)
	vm.DataDir = dir

	fileInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	vm.Volumes = make(map[uint64]*volume.Volume)
	for _, fi := range fileInfos {
		fileName := fi.Name()
		if strings.HasSuffix(fileName, ".data") {
			vid, err := strconv.ParseUint(fileName[:len(fileName) - 5], 10, 64)
			if err != nil {
				panic(err)
			}

			vm.Volumes[vid], err = volume.NewVolume(dir, vid)
			if err != nil {
				panic(err)
			}
		}
	}

	vm.AdminPort = 7800
	vm.AdminHost = "localhost"
	vm.PublicPort = 7900
	vm.PublicHost = "localhost"

	vm.AdminServer = http.NewServeMux()
	vm.PublicServer = http.NewServeMux()
	vm.PublicServer.HandleFunc("/", vm.publicEntry)
	vm.AdminServer.HandleFunc("/", vm.adminEntry)

	vm.MasterHost = "localhost"
	vm.MasterPort = 8888
	return vm, nil
}

func (vm *VolumeManager)Start() {
	go vm.Heartbeat()

	go func() {
		err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", vm.AdminPort), vm.AdminServer)
		if err != nil {
			panic(err)
		}
	}()

	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", vm.PublicPort), vm.PublicServer)
	if err != nil {
		panic(err)
	}
}

func (vm *VolumeManager)Stop() {
	for _, v := range vm.Volumes {
		v.Close()
	}
}

func (vm *VolumeManager)Heartbeat() {
	tick := time.NewTicker(HeartbeatDuration)
	defer tick.Stop()
	for {
		vms := new(master.VolumeManagerStatus)
		vms.AdminHost = vm.AdminHost
		vms.AdminPort = vm.AdminPort
		vms.PublicHost = vm.PublicHost
		vms.PublicPort = vm.PublicPort
		vms.Machine = vm.Machine
		vms.DataCenter = vm.Machine
		vms.VStatusList = make([]*master.VolumeStatus, 0, len(vm.Volumes))

		diskUsage, _ := disk.DiskUsage(vm.DataDir)
		vms.DiskSize = diskUsage.Size
		vms.DiskUsed = diskUsage.Used
		vms.DiskFree = diskUsage.Free

		diskUsedPercent := uint(float64(diskUsage.Used) / float64(diskUsage.Size) * 100)
		if ReadOnly || diskUsedPercent >= MaxDiskUsedPercent {
			//禁止所有volume再进行truncate
			volume.MaxVolumeSize = 0
			vms.CanCreateVolume = false
		} else {
			vms.CanCreateVolume = true
		}

		for vid, v := range vm.Volumes {
			vs := new(master.VolumeStatus)
			vs.Id = vid
			vs.DataFileSize = v.GetDatafileSize()
			vs.Writable = !ReadOnly && v.WriteAble
			vs.MaxFreeSpace = v.GetMaxFreeSpace()
			vms.VStatusList = append(vms.VStatusList, vs)
		}

		api.Heartbeat(vm.MasterHost, vm.MasterPort, vms)
		<-tick.C
	}
}
