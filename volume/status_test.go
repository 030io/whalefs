package volume

import (
	"testing"
	"io/ioutil"
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

	err = status.freeSpace(0, 10000)
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 10; i++ {
		_, err = status.newSpace(1000)
		if err != nil {
			t.Error(err)
		}
	}

	_, err = status.newSpace(1000)
	if err == nil {
		t.Error("err == nil")
	}

	err = status.freeSpace(10000, 10000)
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 10; i++ {
		offset, err = status.newSpace(1000)
		if err != nil {
			t.Error(err)
		}

		err = status.freeSpace(offset, 1000)
		if err != nil {
			t.Error(err)
		}
	}

	_, err = status.newSpace(1000)
	if err != nil {
		t.Error("err == nil")
	}
}