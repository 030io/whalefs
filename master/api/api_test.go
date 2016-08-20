package api_test

import (
	"testing"
	"github.com/030io/whalefs/master"
	"io/ioutil"
	"os"
	"github.com/030io/whalefs/volume/manager"
	"time"
	"github.com/030io/whalefs/master/api"
	volumeApi "github.com/030io/whalefs/volume/api"
	"crypto/rand"
	"crypto/sha1"
)

func TestAPI(t *testing.T) {
	m, err := master.NewMaster()
	if err != nil {
		t.Fatal(err)
	}
	m.Metadata, _ = master.NewMetadataRedis("localhost", 6379, "", 10)

	go m.Start()

	dir, _ := ioutil.TempDir("", "")
	defer os.RemoveAll(dir)

	vm, err := manager.NewVolumeManager(dir)
	if err != nil {
		t.Fatal(err)
	}

	vm.AdminPort = 7801
	vm.PublicPort = 7901
	vm.MasterPort = m.Port

	//test heartbeat
	//vm.Volumes[0], _ = volume.NewVolume(dir, 0)
	go vm.Start()

	//test heartbeat
	time.Sleep(time.Second / 10)

	if len(m.VMStatusList) == 0 {
		t.Error("len(m.VMStatusList) == 0")
	}
	//if len(m.VStatusListMap) == 0 {
	//	t.Errorf("len(m.VStatusListMap) == 0")
	//}

	//test upload
	//for i := 1; i < 101; i ++ {
	//size := i / 10 * 1024
	for _, size := range []int{1, 1024, 1024 * 1024, 1024 * 1024 * 2} {
		for i := 0; i < 3; i++ {
			data := make([]byte, size)
			rand.Read(data)

			tempFile, _ := ioutil.TempFile(os.TempDir(), "")
			tempFile.Write(data)
			tempFile.Close()
			err = api.Upload("localhost", m.Port, tempFile.Name(), tempFile.Name())
			if err != nil {
				t.Fatal(err)
			}

			//test get
			body, err := api.Get("localhost", m.Port, tempFile.Name())
			if err != nil {
				t.Fatal(err)
			}else if sha1.Sum(body) != sha1.Sum(data) {
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
	}
}

func TestReplication(t *testing.T) {
	m, err := master.NewMaster()
	if err != nil {
		t.Fatal(err)
	}
	m.Metadata, _ = master.NewMetadataRedis("localhost", 6379, "", 11)
	m.Port = 7998
	m.PublicPort = 7997
	m.Replication = [3]int{1, 0, 0}
	go m.Start()

	dir, _ := ioutil.TempDir("", "")
	defer os.RemoveAll(dir)

	vm1, err := manager.NewVolumeManager(dir)
	if err != nil {
		t.Fatal(err)
	}

	vm1.AdminPort = 7802
	vm1.PublicPort = 7902
	vm1.MasterPort = m.Port
	go vm1.Start()

	dir2, _ := ioutil.TempDir("", "")
	defer os.RemoveAll(dir2)

	vm2, err := manager.NewVolumeManager(dir2)
	if err != nil {
		t.Fatal(err)
	}

	vm2.AdminPort = 7803
	vm2.PublicPort = 7903
	vm2.MasterPort = m.Port
	go vm2.Start()

	time.Sleep(time.Second / 10)

	if len(m.VMStatusList) != 2 {
		t.Fatalf("len(m.VMStatusList):%d != 2", len(m.VMStatusList))
	}

	for _, size := range []int{1, 1024, 1024 * 1024, 1024 * 1024 * 2} {
		for i := 0; i < 3; i++ {
			data := make([]byte, size)
			rand.Read(data)

			tempFile, _ := ioutil.TempFile(os.TempDir(), "")
			tempFile.Write(data)
			tempFile.Close()
			err = api.Upload("localhost", m.Port, tempFile.Name(), tempFile.Name())
			if err != nil {
				t.Fatal(err)
			}

			//test get
			body, err := api.Get("localhost", m.Port, tempFile.Name())
			if err != nil {
				t.Fatal(err)
			}else if sha1.Sum(body) != sha1.Sum(data) {
				t.Error("data wrong")
			}

			vid, fid, fileName, err := m.Metadata.Get(tempFile.Name())
			if err != nil {
				t.Fatal(err)
			}

			body, err = volumeApi.Get(vm1.PublicHost, vm1.PublicPort, vid, fid, fileName)
			if err != nil {
				t.Fatal(err)
			}else if sha1.Sum(body) != sha1.Sum(data) {
				t.Error("data wrong")
			}

			body, err = volumeApi.Get(vm2.PublicHost, vm2.PublicPort, vid, fid, fileName)
			if err != nil {
				t.Fatal(err)
			}else if sha1.Sum(body) != sha1.Sum(data) {
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
				t.Error("delete failed?")
			}

			_, err = volumeApi.Get(vm1.PublicHost, vm1.PublicPort, vid, fid, fileName)
			if err == nil {
				t.Fatal("delete failed?")
			}

			_, err = volumeApi.Get(vm2.PublicHost, vm2.PublicPort, vid, fid, fileName)
			if err == nil {
				t.Fatal("delete failed?")
			}
		}
	}
}
