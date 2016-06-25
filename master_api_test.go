package main

import (
	"testing"
	"github.com/030io/whalefs/master"
	"io/ioutil"
	"os"
	"github.com/030io/whalefs/volume/manager"
	"github.com/030io/whalefs/volume"
	"time"
)

func TestHeartbeat(t *testing.T) {
	m, err := master.NewMaster()
	if err != nil {
		t.Fatal(err)
	}

	go m.Start()

	dir, _ := ioutil.TempDir("", "")
	defer os.RemoveAll(dir)

	vm, err := manager.NewVolumeManager(dir)
	if err != nil {
		t.Fatal(err)
	}

	vm.AdminPort = 7801
	vm.PublicPort = 7901
	vm.Volumes[0], _ = volume.NewVolume(dir, 0)
	go vm.Start()

	time.Sleep(time.Second / 10)

	if len(m.VMStatusList) == 0 {
		t.Errorf("len(m.VMStatusList) == 0")
	}
	if len(m.VStatusListMap) == 0 {
		t.Errorf("len(m.VStatusListMap) == 0")
	}
}
