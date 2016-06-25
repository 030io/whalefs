package master

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
	"time"
//"net/url"
//"fmt"
	"path/filepath"
)

func (m *Master)masterEntry(w http.ResponseWriter, r *http.Request) {
	m.serverMutex.RLock()
	defer m.serverMutex.RUnlock()

	switch r.URL.Path {
	case "/heartbeat":
		m.heartbeat(w, r)
	default:
		if r.URL.Path == "/favicon.ico" || len(r.URL.Path) <= 1 {
			http.NotFound(w, r)
			return
		}

		switch r.Method{
		case http.MethodGet:
			m.getFile(w, r)
		case http.MethodPost:
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

func (m *Master)getFile(w http.ResponseWriter, r *http.Request) {
	vid, fid, fileName, err := m.Metadata.Get(r.URL.Path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	vStatusList, ok := m.VStatusListMap[vid]
	if !ok {
		http.Error(w, "", http.StatusNotFound)
	}

	//TODO: http 300
	vStatus := vStatusList[0]
	http.Redirect(w, r, vStatus.getFileUrl(fid, fileName), http.StatusFound)
}

func (m *Master)uploadFile(w http.ResponseWriter, r *http.Request) {
	//如果存在则删除旧文件,再上传新文件
	vid, fid, fileName, err := m.Metadata.Get(r.URL.Path)
	if err == nil {
		vStatusList := m.VStatusListMap[vid]
		vStatus := vStatusList[0]
		err = vStatus.delete(fid, fileName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		m.Metadata.Delete(r.URL.Path)
	}

	vStatusList := m.getWritableVolumes()
	if len(vStatusList) == 0 {
		http.Error(w, "not writable volumes", http.StatusInternalServerError)
		return
	}

	//TODO: replicate
	vStatus := vStatusList[0]
	fid = m.generateFid()
	err = vStatus.uploadWithHTTP(r, fid, filepath.Base(r.URL.Path))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	m.Metadata.Set(r.URL.Path, vStatus.Id, fid, filepath.Base(r.URL.Path))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Error(w, "", http.StatusCreated)
}

func (m *Master)deleteFile(w http.ResponseWriter, r *http.Request) {
	vid, fid, fileName, err := m.Metadata.Get(r.URL.Path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	vStatusList, ok := m.VStatusListMap[vid]
	if !ok {
		http.NotFound(w, r)
		return
	}

	vStatus := vStatusList[0]
	//TODO: replicate
	err = vStatus.delete(fid, fileName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}else {
		m.Metadata.Delete(r.URL.Path)
		http.Error(w, "", http.StatusAccepted)
	}
}
