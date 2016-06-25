package api_test

import (
	"testing"
	"github.com/030io/whalefs/master"
	"io/ioutil"
	"os"
	"github.com/030io/whalefs/volume/manager"
	"github.com/030io/whalefs/volume"
	"time"
	"github.com/030io/whalefs/master/api"
)

func TestAPI(t *testing.T) {
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

	//test heartbeat
	time.Sleep(time.Second / 10)

	if len(m.VMStatusList) == 0 {
		t.Errorf("len(m.VMStatusList) == 0")
	}
	if len(m.VStatusListMap) == 0 {
		t.Errorf("len(m.VStatusListMap) == 0")
	}

	//test upload
	m.Metadata, _ = master.NewMetadataRedis("localhost", 6379, "", 0)

	tempFile, _ := ioutil.TempFile(os.TempDir(), "")
	tempFile.WriteString("1234567890")
	tempFile.Close()
	err = api.Upload("localhost", m.Port, tempFile.Name(), "")
	if err != nil {
		t.Fatal(err)
	}

	//test get
	body, err := api.Get("localhost", m.Port, tempFile.Name())
	if err != nil {
		t.Fatal(err)
	}else if string(body) != "1234567890" {
		t.Error("data wrong")
	}

	//test delete
	err = api.Delete("localhost", m.Port, tempFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	//test get again
	_, err = api.Get("localhost", m.Port, tempFile.Name())
	if err == nil {
		t.Error("delete fail?")
	}
}

