package master

import (
	"net/http"
	"fmt"
	"sync"
	"errors"
	"math/rand"
	"github.com/030io/whalefs/utils/uuid"
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

func (m *Master)getWritableVolumes(size uint64) ([]*VolumeStatus, error) {
	m.statusMutex.RLock()
	defer m.statusMutex.RUnlock()

	//map 迭代是随机的,所以不需要手动负载均衡
	for _, vStatusList := range m.VStatusListMap {
		if m.volumesIsValid(vStatusList) && m.volumesIsWritable(vStatusList, size) {
			return vStatusList, nil
		}
	}

	return nil, errors.New("can't find writable volumes")
}

func (m *Master)volumesIsValid(vStatusList []*VolumeStatus) bool {
	for _, vs := range vStatusList {
		if !vs.vmStatus.IsAlive() {
			return false
		}
	}

	if len(vStatusList) != 1 + m.Replication[0] + m.Replication[1] + m.Replication[2] {
		return false
	}

	return true
}

func (m *Master)volumesIsWritable(vStatusList []*VolumeStatus, size uint64) bool {
	for _, vs := range vStatusList {
		if !vs.IsWritable(size) {
			return false
		}
	}
	return len(vStatusList) != 0
}

func (m *Master)vmsNeedCreateVolume(vms *VolumeManagerStatus) bool {
	m.statusMutex.RLock()
	defer m.statusMutex.RUnlock()

	need := true
	for _, vs := range vms.VStatusList {
		if m.volumesIsValid(m.VStatusListMap[vs.Id]) && m.volumesIsWritable(m.VStatusListMap[vs.Id], 0) {
			need = false
		}
	}

	return need && vms.CanCreateVolume
}

func (m *Master)createVolumeWithReplication(vms *VolumeManagerStatus) error {
	if !vms.IsAlive() {
		return fmt.Errorf("%s:%d is dead", vms.AdminHost, vms.AdminPort)
	}

	vmsList, err := m.getMatchReplicationVMS(vms)
	if err != nil {
		return err
	}

	vid := uuid.GenerateUUID()
	for _, vms := range vmsList {
		err := vms.createVolume(vid)
		if err != nil {
			return err
		}
	}

	vStatusList := make([]*VolumeStatus, 0, len(vmsList))
	for _, vms := range vmsList {
		for _, vs := range vms.VStatusList {
			if vs.Id == vid {
				vStatusList = append(vStatusList, vs)
				break
			}
		}
	}
	m.statusMutex.Lock()
	m.VStatusListMap[vid] = vStatusList
	m.statusMutex.Unlock()
	return nil
}

func (m *Master)getMatchReplicationVMS(vms *VolumeManagerStatus) ([]*VolumeManagerStatus, error) {
	m.statusMutex.RLock()
	defer m.statusMutex.RUnlock()

	vmsList := []*VolumeManagerStatus{vms}

	VMStatusList := append(make([]*VolumeManagerStatus, 0, len(m.VMStatusList)), m.VMStatusList...)
	length := len(VMStatusList)
	for i := 0; i < length; i++ {
		a := rand.Intn(length)
		b := rand.Intn(length)
		VMStatusList[a], VMStatusList[b] = VMStatusList[b], VMStatusList[a]
	}

	find0:
	for _, vms := range VMStatusList {
		if len(vmsList) == 1 + m.Replication[0] {
			break
		}
		if vms.IsAlive() && vms.CanCreateVolume {
			for _, vms_ := range vmsList {
				if vms == vms_ || vms.Machine != vms_.Machine || vms.DataCenter != vms_.DataCenter {
					continue find0
				}
			}
			vmsList = append(vmsList, vms)
		}
	}
	if len(vmsList) != 1 + m.Replication[0] {
		return vmsList, errors.New("can't find enough 'same machine VM' to create volume")
	}

	find1:
	for _, vms := range VMStatusList {
		if len(vmsList) == 1 + m.Replication[0] + m.Replication[1] {
			break
		}
		if vms.IsAlive() && vms.CanCreateVolume {
			for _, vms_ := range vmsList {
				if vms == vms_ || vms.Machine == vms_.Machine || vms.DataCenter != vms_.DataCenter {
					continue find1
				}
			}
			vmsList = append(vmsList, vms)
		}
	}
	if len(vmsList) != 1 + m.Replication[0] + m.Replication[1] {
		return vmsList, errors.New("can't find enough 'different machine but same datacenter VM' to create volume")
	}

	find2:
	for _, vms := range VMStatusList {
		if len(vmsList) == 1 + m.Replication[0] + m.Replication[1] + m.Replication[2] {
			break
		}
		if vms.IsAlive() && vms.CanCreateVolume {
			for _, vms_ := range vmsList {
				if vms == vms_ || vms.Machine == vms_.Machine || vms.DataCenter == vms_.DataCenter {
					continue find2
				}
			}
			vmsList = append(vmsList, vms)
		}
	}
	if len(vmsList) != 1 + m.Replication[0] + m.Replication[1] + m.Replication[2] {
		return vmsList, errors.New("can't find enough 'different machine and different datacenter VM' to create volume")
	}

	return vmsList, nil
}

func (m *Master)updateVMS(newVms *VolumeManagerStatus) {
	m.statusMutex.Lock()
	defer m.statusMutex.Unlock()

	for i, oldVms := range m.VMStatusList {
		if oldVms.AdminHost == newVms.AdminHost && oldVms.AdminPort == newVms.AdminPort {
			m.VMStatusList = append(m.VMStatusList[:i], m.VMStatusList[i + 1:]...)
			for _, vs := range oldVms.VStatusList {
				vsList := m.VStatusListMap[vs.Id]
				if len(vsList) == 1 {
					delete(m.VStatusListMap, vs.Id)
					continue
				}
				for i, vs_ := range vsList {
					if vs == vs_ {
						vsList = append(vsList[:i], vsList[i + 1:]...)
						break
					}
				}
				m.VStatusListMap[vs.Id] = vsList
			}
			break
		}
	}

	m.VMStatusList = append(m.VMStatusList, newVms)

	for _, vs := range newVms.VStatusList {
		vs.vmStatus = newVms
		vsList := m.VStatusListMap[vs.Id]
		if vsList == nil {
			vsList = []*VolumeStatus{vs}
		} else {
			vsList = append(vsList, vs)
		}
		m.VStatusListMap[vs.Id] = vsList
	}
}
