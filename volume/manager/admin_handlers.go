package manager

import (
	"net/http"
	"regexp"
	"strconv"
	"io"
	"fmt"
//"encoding/json"
	"github.com/030io/whalefs/volume"
)

var (
	postVolumeUrl *regexp.Regexp
	postFileUrl *regexp.Regexp
	deleteFileUrl  *regexp.Regexp
)

func init() {
	var err error
	postVolumeUrl, err = regexp.Compile("/([0-9]*)/")
	if err != nil {
		panic(err)
	}

	postFileUrl, err = regexp.Compile("/([0-9]*)/([0-9]*)/(.*)")
	if err != nil {
		panic(err)
	}

	deleteFileUrl = postFileUrl
}

type Size interface {
	Size() int64
}

func (vm *VolumeManager)adminEntry(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		if postFileUrl.MatchString(r.URL.Path) {
			vm.adminPostFile(w, r)
		}else if postVolumeUrl.MatchString(r.URL.Path) {
			vm.adminPostVolume(w, r)
		}else {
			http.Error(w, r.URL.String() + " can't match", http.StatusNotFound)
		}
	case http.MethodDelete:
		if deleteFileUrl.MatchString(r.URL.Path) {
			vm.adminDeleteFile(w, r)
		}else {
			http.Error(w, r.URL.String() + " can't match", http.StatusNotFound)
		}
	default:
		http.Error(w, "", http.StatusMethodNotAllowed)
	}
}

func (vm *VolumeManager)adminPostVolume(w http.ResponseWriter, r *http.Request) {
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

func (vm *VolumeManager)adminPostFile(w http.ResponseWriter, r *http.Request) {
	match := postFileUrl.FindStringSubmatch(r.URL.Path)
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

	fileSize := file.(Size).Size()

	fid, _ := strconv.ParseUint(match[2], 10, 64)
	file_, err := volume.NewFile(fid, header.Filename, uint64(fileSize))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err != nil {
			if err := volume.Delete(file_.Info.Fid, file_.Info.FileName); err != nil {
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

	w.WriteHeader(http.StatusCreated)
}

func (vm *VolumeManager)adminDeleteFile(w http.ResponseWriter, r *http.Request) {
	match := deleteFileUrl.FindStringSubmatch(r.URL.Path)
	vid, _ := strconv.Atoi(match[1])
	volume := vm.Volumes[vid]
	if volume == nil {
		http.Error(w, "can't find volume", http.StatusNotFound)
		return
	}

	fid, _ := strconv.ParseUint(match[2], 10, 64)
	err := volume.Delete(fid, match[3])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}else {
		http.Error(w, "", http.StatusAccepted)
	}
}
