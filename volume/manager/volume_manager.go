package manager

import (
	"net/http"
	"github.com/030io/whalefs/volume"
	"os"
	"io/ioutil"
	"strings"
	"strconv"
	"fmt"
)

type VolumeManager struct {
	DataDir      string
	Volumes      map[int]*volume.Volume
	AdminPort    int
	AdminHost    string
	PublicPort   int
	PublicHost   string
	AdminServer  *http.ServeMux
	PublicServer *http.ServeMux
}

func NewVolumeManager(dir string) (*VolumeManager, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) || os.IsPermission(err) {
		panic(err)
	}

	vm := new(VolumeManager)
	vm.DataDir = dir

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

	vm.Volumes = make(map[int]*volume.Volume)

	vm.AdminPort = 7800
	vm.AdminHost = "localhost"
	vm.PublicPort = 7900
	vm.PublicHost = "localhost"

	vm.AdminServer = http.NewServeMux()
	vm.PublicServer = http.NewServeMux()
	vm.PublicServer.HandleFunc("/", vm.publicEntry)
	vm.AdminServer.HandleFunc("/", vm.adminEntry)
	return vm, nil
}

func (vm *VolumeManager)Start() {
	go func() {
		err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", vm.AdminPort), vm.AdminServer)
		if err != nil {
			panic(err)
		}
	}()

	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", vm.PublicPort), vm.PublicServer)
	if err != nil {
		panic(err)
	}
}

func (vm *VolumeManager)Stop() {
	for _, v := range vm.Volumes {
		v.Close()
	}
}
