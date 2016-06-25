package master

import (
	"net/http"
	"fmt"
	"sync"
	"math/rand"
	"strconv"
)

const vidKey = "__vid"
const fidKey = "__fid"

type Master struct {
	Port           int

	VMStatusList   []*VolumeManagerStatus
	VStatusListMap map[int][]*VolumeStatus

	Server         *http.ServeMux
	serverMutex    sync.RWMutex

	Metadata       Metadata
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

func (m *Master)getWritableVolumes() []*VolumeStatus {
	for _, vStatusList := range m.VStatusListMap {
		return vStatusList
	}
	return nil
}

func (m *Master)generateFid() uint64 {
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
	return uint64(rand.Uint32())
}
