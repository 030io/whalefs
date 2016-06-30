package master

import (
	"net/http"
	"fmt"
	"sync"
	"strconv"
	"github.com/syndtr/goleveldb/leveldb/errors"
)

const (
	vidKey = "__vid"
	fidKey = "__fid"
)

type Master struct {
	Port           int

	VMStatusList   []*VolumeManagerStatus
	VStatusListMap map[int][]*VolumeStatus
	volumeMutex    sync.RWMutex

	Server         *http.ServeMux
	serverMutex    sync.RWMutex

	fidMutex       sync.Mutex
	vidMutex       sync.Mutex

	Metadata       Metadata
	Replication    [3]int
}

func NewMaster() (*Master, error) {
	m := new(Master)
	m.Port = 8888
	m.VMStatusList = make([]*VolumeManagerStatus, 0, 1)
	m.VStatusListMap = make(map[int][]*VolumeStatus)

	m.Server = http.NewServeMux()
	m.Server.HandleFunc("/", m.masterEntry)

	m.Replication = [3]int{
		0, //number of replica in the same machine
		0, //number of replica in the different machine
		0, //number of replica in the different datacenter
	}
	return m, nil
}

func (m *Master)Start() {
	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", m.Port), m.Server)
	if err != nil {
		panic(err)
	}
}

func (m *Master)Stop() {
	m.serverMutex.Lock()
}

func (m *Master)getWritableVolumes() ([]*VolumeStatus, error) {
	m.volumeMutex.Lock()
	defer m.volumeMutex.Unlock()

	//TODO: 优化负载均衡

	for _, vStatusList := range m.VStatusListMap {
		if m.vStatusListIsValid(vStatusList) {
			return vStatusList, nil
		}
	}

	err := m.createVolumeWithReplication()
	if err != nil {
		return nil, err
	}

	for _, vStatusList := range m.VStatusListMap {
		if m.vStatusListIsValid(vStatusList) {
			return vStatusList, nil
		}
	}

	return nil, errors.New("can't find writable volumes")
}

func (m *Master)vStatusListIsValid(vStatusList []*VolumeStatus) bool {
	for _, vs := range vStatusList {
		if !vs.vmStatus.IsAlive() {
			return false
		}
	}

	if len(vStatusList) != 1 + m.Replication[0] + m.Replication[1] + m.Replication[2] {
		return false
	}

	//TODO: check volume writable
	return true
}

func (m *Master)createVolumeWithReplication() error {
	temp := make([]*VolumeManagerStatus, 0)
	for _, vms := range m.VMStatusList {
		if vms.IsAlive() && vms.canCreateVolume() {
			temp = append(temp, vms)
			break
		}
	}

	find0:
	for _, vms := range m.VMStatusList {
		if len(temp) == 1 + m.Replication[0] {
			break
		}
		if vms.IsAlive() && vms.canCreateVolume() {
			for _, vms_ := range temp {
				if vms == vms_ || vms.Machine != vms_.Machine || vms.DataCenter != vms_.DataCenter {
					continue find0
				}
			}
			temp = append(temp, vms)
		}
	}
	if len(temp) != 1 + m.Replication[0] {
		return errors.New("can't find enough 'same machine VM' to create volume")
	}

	find1:
	for _, vms := range m.VMStatusList {
		if len(temp) == 1 + m.Replication[0] + m.Replication[1] {
			break
		}
		if vms.IsAlive() && vms.canCreateVolume() {
			for _, vms_ := range temp {
				if vms == vms_ || vms.Machine == vms_.Machine || vms.DataCenter != vms_.DataCenter {
					continue find1
				}
			}
			temp = append(temp, vms)
		}
	}
	if len(temp) != 1 + m.Replication[0] + m.Replication[1] {
		return errors.New("can't find enough 'different machine but same datacenter VM' to create volume")
	}

	find2:
	for _, vms := range m.VMStatusList {
		if len(temp) == 1 + m.Replication[0] + m.Replication[1] + m.Replication[2] {
			break
		}
		if vms.IsAlive() && vms.canCreateVolume() {
			for _, vms_ := range temp {
				if vms == vms_ || vms.Machine == vms_.Machine || vms.DataCenter == vms_.DataCenter {
					continue find2
				}
			}
			temp = append(temp, vms)
		}
	}
	if len(temp) != 1 + m.Replication[0] + m.Replication[1] + m.Replication[2] {
		return errors.New("can't find enough 'different machine and different datacenter VM' to create volume")
	}

	vid := m.generateVid()
	for _, vms := range temp {
		err := vms.createVolume(vid)
		if err != nil {
			return err
		}

		m.VStatusListMap[vid] = append(m.VStatusListMap[vid], vms.VStatusList[len(vms.VStatusList) - 1])
	}
	return nil
}

func (m *Master)generateVid() int {
	m.vidMutex.Lock()
	defer m.vidMutex.Unlock()

	value, err := m.Metadata.getConfig(vidKey)
	if err != nil {
		value = "0"
	}

	vid, _ := strconv.Atoi(value)
	for i := 0; i < 3; i++ {
		err = m.Metadata.setConfig(vidKey, strconv.Itoa(vid + 1))
		if err == nil {
			break
		}
	}
	if err != nil {
		panic(err)
	}

	return vid
}

func (m *Master)generateFid() uint64 {
	m.fidMutex.Lock()
	defer m.fidMutex.Unlock()

	value, err := m.Metadata.getConfig(fidKey)
	if err != nil {
		value = "0"
	}
	fid, _ := strconv.ParseUint(value, 10, 64)
	for i := 0; i < 3; i++ {
		err = m.Metadata.setConfig(fidKey, strconv.FormatUint(fid + 1, 10))
		if err == nil {
			break
		}
	}
	if err != nil {
		panic(err)
	}

	return fid
}
