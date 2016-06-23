package api

import (
	"testing"
	"io/ioutil"
	"os"
	"github.com/030io/whalefs/volume/server"
)

func TestAPI(t *testing.T) {
	dir, _ := ioutil.TempDir("", "")
	defer os.RemoveAll(dir)

	vm, err := server.NewVolumeManager(dir)
	if err != nil {
		t.Error(err)
		return
	}

	vm.AdminPort = 7800
	vm.PublicPort = 7900

	go vm.Start()

	//test create volume
	err = CreateVolume(vm.AdminHost, vm.AdminPort, 1)
	if err != nil {
		t.Error(err)
	}

	testData := []byte("1234567890")

	tempFile, _ := ioutil.TempFile(os.TempDir(), "test_api_")
	tempFile.Write(testData)
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	//test upload
	vid, fid, filename, err := Upload(vm.AdminHost, vm.AdminPort, 1, tempFile.Name())
	if err != nil {
		t.Error(err)
	}else if vid != 1 || fid != 0 || len(filename) == 0 {
		t.Error(vid, fid, filename)
	}

	//test get file
	data, err := GetFile(vm.PublicHost, vm.PublicPort, vid, fid, filename)
	if err != nil {
		t.Error(err)
	}

	if string(data) != string(testData) {
		t.Errorf("%s != %s", data, testData)
	}

	//test delete file
	err = DeleteFile(vm.AdminHost, vm.AdminPort, vid, fid, filename)
	if err != nil {
		t.Error(err)
	}

	_, err = GetFile(vm.PublicHost, vm.PublicPort, vid, fid, filename)
	if err == nil {
		t.Error(err)
	}
}
