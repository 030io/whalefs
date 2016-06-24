package api

import (
	"testing"
	"io/ioutil"
	"os"
	"github.com/030io/whalefs/volume/manager"
	"path/filepath"
)

func TestAPI(t *testing.T) {
	dir, _ := ioutil.TempDir("", "")
	defer os.RemoveAll(dir)

	vm, err := manager.NewVolumeManager(dir)
	if err != nil {
		t.Error(err)
		return
	}

	vm.AdminPort = 7800
	vm.PublicPort = 7900

	go vm.Start()

	testVid := 1
	//test create volume
	err = CreateVolume(vm.AdminHost, vm.AdminPort, testVid)
	if err != nil {
		t.Error(err)
	}

	testData := []byte("1234567890")

	tempFile, _ := ioutil.TempFile(os.TempDir(), "test_api_")
	tempFile.Write(testData)
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	//test upload
	testFid := uint64(1)
	fileName := filepath.Base(tempFile.Name())
	err = Upload(vm.AdminHost, vm.AdminPort, testVid, testFid, tempFile.Name(), "")
	if err != nil {
		t.Error(err)
	}

	//test get file
	data, err := GetFile(vm.PublicHost, vm.PublicPort, testVid, testFid, fileName)
	if err != nil {
		t.Error(err)
	}

	if string(data) != string(testData) {
		t.Errorf("%s != %s", data, testData)
	}

	//test delete file
	err = DeleteFile(vm.AdminHost, vm.AdminPort, testVid, testFid, fileName)
	if err != nil {
		t.Error(err)
	}

	_, err = GetFile(vm.PublicHost, vm.PublicPort, testVid, testFid, fileName)
	if err == nil {
		t.Error(err)
	}
}
