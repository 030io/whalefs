package volume

import (
	"testing"
	"io/ioutil"
	"encoding/binary"
)

func TestStatus(t *testing.T) {
	dir, _ := ioutil.TempDir("", "test_status_")

	status, err := NewStatus(dir, 0)
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 1000; i++ {
		fid := status.newFid()
		if fid != uint64(i) {
			t.Errorf("fid: %d != i: %d", fid, i)
		}
	}

	offset, err := status.newSpace(10)
	if err == nil {
		t.Error("err == nil")
	}
	if offset != 0 {
		t.Errorf("%d != 0", offset)
	}

	err = status.freeSpace(100, 10000)
	if err != nil {
		t.Error(err)
	}

	value, err := status.db.Get(append(freeSpacePrefix, 0, 0, 0, 0, 0, 0, 0, 100), nil)
	if err != nil {
		t.Error(err)
	}else if binary.BigEndian.Uint64(value) != 10000 {
		t.Error("binary.BigEndian.Uint64(value) != 10000")
	}

	offset, err = status.newSpace(1000)
	if offset != 100 {
		t.Errorf("%d != 100", offset)
	}

	offset, err = status.newSpace(1000)
	if offset != 1100 {
		t.Errorf("%d != 1100", offset)
	}

	err = status.freeSpace(100, 1000)
	if err != nil {
		t.Error(err)
	}

	offset, err = status.newSpace(1000)
	if offset != 100 {
		t.Errorf("%d != 100", offset)
	}
}