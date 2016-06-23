package server

import (
	"net/http"
	"regexp"
	"strconv"
	"io"
	"fmt"
	"encoding/json"
	"github.com/030io/whalefs/volume"
)

var (
	postVolumeUrl *regexp.Regexp
	putFileUrl *regexp.Regexp
	deleteFileUrl  *regexp.Regexp
)

func init() {
	var err error
	postVolumeUrl, err = regexp.Compile("/([0-9]*)/")
	if err != nil {
		panic(err)
	}

	putFileUrl = postVolumeUrl

	deleteFileUrl, err = regexp.Compile("/([0-9]*)/([0-9]*)/(.*)")
	if err != nil {
		panic(err)
	}
}

type Size interface {
	Size() int64
}

func (vm *VolumeManager)adminEntry(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		vm.adminPostVolume(w, r)
	case http.MethodPut:
		vm.adminPutFile(w, r)
	case http.MethodDelete:
		vm.adminDeleteFile(w, r)
	default:
		http.Error(w, "", http.StatusMethodNotAllowed)
	}
}

func (vm *VolumeManager)adminPostVolume(w http.ResponseWriter, r *http.Request) {
	if postVolumeUrl.MatchString(r.URL.Path) == false {
		http.Error(w, r.URL.String() + " can't match", http.StatusNotFound)
		return
	}

	match := postVolumeUrl.FindStringSubmatch(r.URL.Path)
	vid, _ := strconv.Atoi(match[1])
	volume, err := volume.NewVolume(vm.DataDir, vid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vm.Volumes[vid] = volume
	http.Error(w, "", http.StatusCreated)
}

func (vm *VolumeManager)adminPutFile(w http.ResponseWriter, r *http.Request) {
	if putFileUrl.MatchString(r.URL.Path) == false {
		http.Error(w, r.URL.String() + " can't match", http.StatusNotFound)
		return
	}

	match := putFileUrl.FindStringSubmatch(r.URL.Path)
	vid, _ := strconv.Atoi(match[1])
	volume := vm.Volumes[vid]
	if volume == nil {
		http.Error(w, "can't find volume", http.StatusNotFound)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	//test
	//if _, ok := file.(Size); ok {
	//	panic(ok)
	//}else {
	//	panic(ok)
	//}

	fileSize := file.(Size).Size()
	//fileSize, err := file.Seek(0, 2)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}
	//_, err = file.Seek(0, 0)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}

	file_, err := volume.NewFile(header.Filename, uint64(fileSize))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err != nil {
			if err := volume.Delete(file_.Info.Fid); err != nil {
				panic(err)
			}
		}
	}()

	n, err := io.Copy(file_, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}else if n != fileSize {
		panic(fmt.Errorf("%d != %d", n, fileSize))
	}

	//TODO: replicate
	result := &PostFileResult{
		Vid: vid,
		Fid: file_.Info.Fid,
		FileName: file_.Info.FileName,
	}
	data, _ := json.Marshal(result)
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (vm *VolumeManager)adminDeleteFile(w http.ResponseWriter, r *http.Request) {
	//TODO: delete
	if deleteFileUrl.MatchString(r.URL.Path) == false {
		http.Error(w, r.URL.String() + " can't match", http.StatusNotFound)
		return
	}

	match := deleteFileUrl.FindStringSubmatch(r.URL.Path)
	vid, _ := strconv.Atoi(match[1])
	volume := vm.Volumes[vid]
	if volume == nil {
		http.Error(w, "can't find volume", http.StatusNotFound)
		return
	}

	fid, _ := strconv.ParseUint(match[2], 10, 64)
	err := volume.Delete(fid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}else {
		http.Error(w, "", http.StatusAccepted)
	}
}
