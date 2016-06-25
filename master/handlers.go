package master

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
	"time"
)

func (m *Master)masterEntry(w http.ResponseWriter, r *http.Request) {
	m.serverMutex.RLock()
	defer m.serverMutex.RUnlock()

	switch r.URL.Path {
	case "/heartbeat":
		m.heartbeat(w, r)
	default:
		switch r.Method{
		case http.MethodPut, http.MethodPost:
			m.uploadFile(w, r)
		case http.MethodDelete:
			m.deleteFile(w, r)
		default:
			http.Error(w, "", http.StatusMethodNotAllowed)
		}
	}
}

func (m *Master)heartbeat(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	newVms := new(VolumeManagerStatus)
	json.Unmarshal(body, newVms)
	newVms.LastHeartbeat = time.Now()

	m.serverMutex.RUnlock()
	defer m.serverMutex.RLock()
	m.serverMutex.Lock()
	defer m.serverMutex.Unlock()

	for i, oldVms := range m.VMStatusList {
		if oldVms.AdminHost == newVms.AdminHost && oldVms.AdminPort == newVms.AdminPort {
			m.VMStatusList = append(m.VMStatusList[:i], m.VMStatusList[i + 1:]...)
			for _, vs := range oldVms.VStatusList {
				vsList := m.VStatusListMap[vs.Id]
				for i, vs_ := range vsList {
					if vs.Id == vs_.Id {
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
	for vid, vs := range newVms.VStatusList {
		vs.vmStatus = newVms
		vsList := m.VStatusListMap[vid]
		if vsList == nil {
			vsList = make([]*VolumeStatus, 0)
		}
		vsList = append(vsList, vs)
		m.VStatusListMap[vid] = vsList
	}
}

//TODO: master upload file
func (m *Master)uploadFile(w http.ResponseWriter, r *http.Request) {
}

//TODO: master delete file
func (m *Master)deleteFile(w http.ResponseWriter, r *http.Request) {

}
