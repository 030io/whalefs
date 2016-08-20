package master

import (
	"net/http"
	"fmt"
	"sync"
	"errors"
	"math/rand"
	"time"
)

type Master struct {
	Port           int
	PublicPort     int

	VMStatusList   []*VolumeManagerStatus
	VStatusListMap map[uint64][]*VolumeStatus
	statusMutex    sync.RWMutex

	Server         *http.ServeMux
	serverMutex    sync.RWMutex
	PublicServer   *http.ServeMux

	Metadata       Metadata
	Replication    [3]int
}

func NewMaster() (*Master, error) {
	m := new(Master)
	m.Port = 8888
	m.PublicPort = 8899
	m.VMStatusList = make([]*VolumeManagerStatus, 0, 1)
	m.VStatusListMap = make(map[uint64][]*VolumeStatus)

	m.Server = http.NewServeMux()
	m.Server.HandleFunc("/", m.masterEntry)
	m.PublicServer = http.NewServeMux()
	m.PublicServer.HandleFunc("/", m.publicEntry)

	m.Replication = [3]int{
		0, //number of replica in the same machine
		0, //number of replica in the different machine
		0, //number of replica in the different datacenter
	}
	return m, nil
}

func (m *Master)Start() {
	go func() {
		err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", m.PublicPort), m.PublicServer)
		if err != nil {
			panic(err)
		}
	}()

	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", m.Port), m.Server)
	if err != nil {
		panic(err)
	}
}

func (m *Master)Stop() {
	m.serverMutex.Lock()
	m.Metadata.Close()
}

func (m *Master)getWritableVolumes() ([]*VolumeStatus, error) {
	m.statusMutex.RLock()
	defer m.statusMutex.RUnlock()

	//map 迭代是随机的,所以不需要手动负载均衡
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

func (m *Master)createVolumeWithReplication(vms *VolumeManagerStatus) error {
	m.statusMutex.RLock()
	defer m.statusMutex.RUnlock()

	if !vms.IsAlive() {
		return fmt.Errorf("%s:%d is dead", vms.AdminHost, vms.AdminPort)
		//}else if !vms.canCreateVolume() {
		//	return fmt.Errorf("%s:%d can't create volume", vms.AdminHost, vms.AdminPort)
	}

	temp := []*VolumeManagerStatus{vms}

	VMStatusList := append(make([]*VolumeManagerStatus, 0, len(m.VMStatusList)), m.VMStatusList...)
	length := len(VMStatusList)
	for i := 0; i < length; i++ {
		a := rand.Intn(length)
		b := rand.Intn(length)
		VMStatusList[a], VMStatusList[b] = VMStatusList[b], VMStatusList[a]
	}

	find0:
	for _, vms := range VMStatusList {
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
	for _, vms := range VMStatusList {
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
	for _, vms := range VMStatusList {
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
	}

	vStatusList := make([]*VolumeStatus, 0, len(temp))
	for _, vms := range temp {
		for _, vs := range vms.VStatusList {
			if vs.Id == vid {
				vStatusList = append(vStatusList, vs)
				break
			}
		}
	}
	m.statusMutex.RUnlock()
	m.statusMutex.Lock()
	m.VStatusListMap[vid] = vStatusList
	m.statusMutex.Unlock()
	m.statusMutex.RLock()

	return nil
}

func (m *Master)generateVid() uint64 {
	return uint64(time.Now().UnixNano())
}

func (m *Master)generateFid() uint64 {
	return uint64(time.Now().UnixNano())
}
