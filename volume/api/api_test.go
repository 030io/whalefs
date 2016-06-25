package api_test

import (
	"testing"
	"io/ioutil"
	"os"
	"github.com/030io/whalefs/volume/manager"
	"github.com/030io/whalefs/volume/api"
	"path/filepath"
)

func TestAPI(t *testing.T) {
	dir, _ := ioutil.TempDir("", "")
	defer os.RemoveAll(dir)

	vm, err := manager.NewVolumeManager(dir)
	if err != nil {
		t.Fatal(err)
	}

	go vm.Start()

	testVid := 1
	//test create volume
	err = api.CreateVolume(vm.AdminHost, vm.AdminPort, testVid)
	if err != nil {
		t.Fatal(err)
	}

	testData := []byte("1234567890")

	tempFile, _ := ioutil.TempFile(os.TempDir(), "test_api_")
	tempFile.Write(testData)
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	//test upload
	testFid := uint64(1)
	fileName := filepath.Base(tempFile.Name())
	err = api.Upload(vm.AdminHost, vm.AdminPort, testVid, testFid, tempFile.Name(), "")
	if err != nil {
		t.Fatal(err)
	}

	//test get file
	data, err := api.GetFile(vm.PublicHost, vm.PublicPort, testVid, testFid, fileName)
	if err != nil {
		t.Error(err)
	}

	if string(data) != string(testData) {
		t.Errorf("%s != %s", data, testData)
	}

	//test delete file
	err = api.DeleteFile(vm.AdminHost, vm.AdminPort, testVid, testFid, fileName)
	if err != nil {
		t.Error(err)
	}

	_, err = api.GetFile(vm.PublicHost, vm.PublicPort, testVid, testFid, fileName)
	if err == nil {
		t.Error(err)
	}
}

