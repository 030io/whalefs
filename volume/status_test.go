package volume

import (
	"testing"
	"io/ioutil"
	"encoding/binary"
	"os"
)

func TestStatus(t *testing.T) {
	dir, _ := ioutil.TempDir("", "test_status_")
	defer os.RemoveAll(dir)

	status, err := NewStatus(dir, 0)
	if err != nil {
		t.Error(err)
	}

	//for i := 0; i < 1000; i++ {
	//	fid := status.newFid()
	//	if fid != uint64(i) {
	//		t.Errorf("fid: %d != i: %d", fid, i)
	//	}
	//}

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

	binary.BigEndian.PutUint64(offsetSize[1:9], 0)
	binary.BigEndian.PutUint64(offsetSize[9:], 10000)
	_, err = status.db.Get(offsetSize, nil)
	if err != nil {
		t.Error(err)
	}

	binary.BigEndian.PutUint64(reversedsizeOffset[1:9], 10000 ^ (^uint64(0)))
	binary.BigEndian.PutUint64(reversedsizeOffset[9:], 0)
	_, err = status.db.Get(reversedsizeOffset, nil)
	if err != nil {
		t.Error(err)
	}

	_, err = status.newSpace(1000)
	if err != nil {
		t.Error(err)
	}

	binary.BigEndian.PutUint64(offsetSize[1:9], 1000)
	binary.BigEndian.PutUint64(offsetSize[9:], 9000)
	_, err = status.db.Get(offsetSize, nil)
	if err != nil {
		t.Error(err)
	}

	binary.BigEndian.PutUint64(reversedsizeOffset[1:9], 9000 ^ (^uint64(0)))
	binary.BigEndian.PutUint64(reversedsizeOffset[9:], 1000)
	_, err = status.db.Get(reversedsizeOffset, nil)
	if err != nil {
		t.Error(err)
	}

	err = status.freeSpace(10000, 10000)
	if err != nil {
		t.Error(err)
	}

	binary.BigEndian.PutUint64(offsetSize[1:9], 1000)
	binary.BigEndian.PutUint64(offsetSize[9:], 19000)
	_, err = status.db.Get(offsetSize, nil)
	if err != nil {
		t.Error(err)
	}

	binary.BigEndian.PutUint64(reversedsizeOffset[1:9], 19000 ^ (^uint64(0)))
	binary.BigEndian.PutUint64(reversedsizeOffset[9:], 1000)
	_, err = status.db.Get(reversedsizeOffset, nil)
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

	binary.BigEndian.PutUint64(offsetSize[1:9], 1000)
	binary.BigEndian.PutUint64(offsetSize[9:], 19000)
	_, err = status.db.Get(offsetSize, nil)
	if err != nil {
		t.Error(err)
	}

	binary.BigEndian.PutUint64(reversedsizeOffset[1:9], 19000 ^ (^uint64(0)))
	binary.BigEndian.PutUint64(reversedsizeOffset[9:], 1000)
	_, err = status.db.Get(reversedsizeOffset, nil)
	if err != nil {
		t.Error(err)
	}

	_, err = status.newSpace(1000)
	if err != nil {
		t.Error("err == nil")
	}
}