package api

import (
	"net/http"
	"github.com/030io/whalefs/volume"
	"os"
	"io/ioutil"
	"strings"
	"strconv"
)

type VolumeManager struct {
	Volumes      map[int]*volume.Volume
	AdminServer  *http.ServeMux
	PublicServer *http.ServeMux
}

func NewVolumeManager(dir string) (*VolumeManager, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) || os.IsPermission(err) {
		panic(err)
	}

	vm := new(VolumeManager)
	vm.AdminServer = http.NewServeMux()
	vm.PublicServer = http.NewServeMux()

	fileInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	for _, fi := range fileInfos {
		fileName := fi.Name()
		if strings.HasSuffix(fileName, ".data") {
			vid, err := strconv.Atoi(fileName[:len(fileName) - 5])
			if err != nil {
				panic(err)
			}

			vm.Volumes[vid], err = volume.NewVolume(dir, vid)
		}
	}

	return vm, nil
}

//TODO: init handlers
func (vm *VolumeManager)initHandlers() {
	vm.PublicServer.HandleFunc("/", vm.publicEntry)
}
