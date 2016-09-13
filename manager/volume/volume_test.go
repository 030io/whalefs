package volume

import (
	"testing"
	"io/ioutil"
	"crypto/sha1"
	"os"
	"crypto/rand"
)

func TestVolumeAndFile(t *testing.T) {
	dir, _ := ioutil.TempDir("", "whalefs_test_volume_")
	defer os.RemoveAll(dir)

	v, err := NewVolume(dir, 0)
	if err != nil {
		t.Error(err)
	}

	for i, size := range []uint64{1, 100, 1024, 1024 * 1024, 1024 * 1024 * 10} {
		file, err := v.NewFile(uint64(i), "test_file.1", size)
		if err != nil {
			t.Error(err)
		}

		data := make([]byte, size)
		rand.Read(data)
		_, err = file.Write(data)
		if err != nil {
			t.Error(err)
		}

		file2, err := v.Get(file.Info.Fid)
		if err != nil {
			t.Error(err)
		}

		data2 := make([]byte, size)
		file2.Read(data2)

		if sha1.Sum(data) != sha1.Sum(data2) {
			t.Error("data wrong")
		}

		err = v.Delete(file.Info.Fid, "test_file.1")
		if err != nil {
			t.Error(err)
		}

		file3, err := v.Get(file.Info.Fid)
		if err == nil || file3 != nil {
			t.Error("delete failed?")
		}
	}
}