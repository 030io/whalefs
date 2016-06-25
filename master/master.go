package master

import (
	"net/http"
	"fmt"
	"sync"
)

type Master struct {
	Port           int

	VMStatusList   []*VolumeManagerStatus
	VStatusListMap map[int][]*VolumeStatus

	Server         *http.ServeMux
	serverMutex    sync.RWMutex
}

func NewMaster() (*Master, error) {
	m := new(Master)
	m.Port = 7999
	m.VMStatusList = make([]*VolumeManagerStatus, 0, 1)
	m.VStatusListMap = make(map[int][]*VolumeStatus)

	m.Server = http.NewServeMux()
	m.Server.HandleFunc("/", m.masterEntry)
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

////TODO: check replication
//func (m *Master)VolumeIsWritable(vid int) bool {
//	vStatusList := m.VStatusListMap[vid]
//	if vStatusList == nil {
//		return false
//	}
//
//	for _, vStatus := range vStatusList {
//		if !vStatus.Writable || !vStatus.vmStatus.IsAlive() {
//			return false
//		}
//	}
//
//	return true
//}
