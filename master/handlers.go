package master

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
	"time"
	"strings"
	"sync"
	"math/rand"
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
		case http.MethodGet, http.MethodHead:
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

	remoteIP := r.RemoteAddr[:strings.LastIndex(r.RemoteAddr, ":")]
	if newVms.AdminHost == "" || newVms.AdminHost == "localhost" {
		newVms.AdminHost = remoteIP
	}
	if newVms.PublicHost == "" || newVms.PublicHost == "localhost" {
		newVms.PublicHost = remoteIP
	}
	if newVms.Machine == "" {
		newVms.Machine = remoteIP
	}

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
	for _, vs := range newVms.VStatusList {
		vs.vmStatus = newVms
		vsList := m.VStatusListMap[vs.Id]
		if vsList == nil {
			vsList = make([]*VolumeStatus, 0)
		}
		vsList = append(vsList, vs)
		m.VStatusListMap[vs.Id] = vsList
	}
}

func (m *Master)getFile(w http.ResponseWriter, r *http.Request) {
	vid, fid, fileName, err := m.Metadata.Get(r.URL.Path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	m.volumeMutex.RLock()
	vStatusList, ok := m.VStatusListMap[vid]
	m.volumeMutex.RUnlock()
	if !ok {
		http.Error(w, "can't find volume", http.StatusNotFound)
		return
	}

	length := len(vStatusList)
	j := rand.Intn(length)
	for i := 0; i < length; i++ {
		vStatus := vStatusList[(i + j) % length]
		if vStatus.vmStatus.IsAlive() {
			http.Redirect(w, r, vStatus.getFileUrl(fid, fileName), http.StatusFound)
			return
		}
	}

	http.Error(w, "all volumes is dead", http.StatusInternalServerError)
}

func (m *Master)uploadFile(w http.ResponseWriter, r *http.Request) {
	//如果存在则删除旧文件,再上传新文件
	//vid, fid, fileName, err := m.Metadata.Get(r.URL.Path)
	//if err == nil {
	//	vStatusList := m.VStatusListMap[vid]
	//	vStatus := vStatusList[0]
	//	err = vStatus.delete(fid, fileName)
	//	if err != nil {
	//		http.Error(w, err.Error(), http.StatusInternalServerError)
	//		return
	//	}
	//	m.Metadata.Delete(r.URL.Path)
	//}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var dst string
	if r.URL.Path[len(r.URL.Path) - 1] == '/' {
		dst = r.URL.Path + filepath.Base(header.Filename)
	}else {
		dst = r.URL.Path
	}
	fileName := filepath.Base(dst)

	if m.Metadata.Has(dst) {
		http.Error(w, "file is existed, you should delete it at first.", http.StatusNotAcceptable)
		return
	}

	vStatusList, err := m.getWritableVolumes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, _ := ioutil.ReadAll(file)
	fid := m.generateFid()

	wg := sync.WaitGroup{}
	for _, vStatus := range vStatusList {
		wg.Add(1)
		go func(vs *VolumeStatus) {
			e := vs.uploadFile(fid, fileName, data)
			if e != nil {
				err = e
			}
			wg.Done()
		}(vStatus)
	}
	wg.Wait()

	if err != nil {
		for _, vStatus := range vStatusList {
			go vStatus.delete(fid, fileName)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	m.Metadata.Set(dst, vStatusList[0].Id, fid, fileName)
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
		http.Error(w, "can't find volume", http.StatusNotFound)
		return
	}else if !m.vStatusListIsValid(vStatusList) {
		http.Error(w, "can't delete file, because it's(volumes) readonly.", http.StatusNotAcceptable)
	}

	for _, vStatus := range vStatusList {
		go vStatus.delete(fid, fileName)
	}

	m.Metadata.Delete(r.URL.Path)
	http.Error(w, "", http.StatusAccepted)
}
