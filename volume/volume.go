package volume

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"encoding/binary"
	"time"
)

type Volume struct {
	MaxSize  uint64

	index    Index
	dataFile *os.File
	rwMutex  sync.RWMutex
}

func NewVolume(dir string, vid int) (v *Volume, err error) {
	path := filepath.Join(dir, strconv.Itoa(vid) + ".data")
	v.dataFile, err = os.OpenFile(path, os.O_CREATE | os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	v.index, err = NewLevelDBIndex(dir, vid)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (v *Volume)Get(fid uint64) (*File, error) {
	v.rwMutex.RLock()
	defer v.rwMutex.RUnlock()

	fi, err := v.index.Get(fid)
	if err == nil {
		return &File{volume: v, Info: fi}, nil
	}else {
		return nil, err
	}
}

//func (v *Volume)Add(*File) error {
//	return nil
//}

func (v *Volume)Delete(fid uint64) error {
	v.rwMutex.Lock()
	defer v.rwMutex.Unlock()

	file, err := v.Get(fid)
	if err != nil {
		return err
	}

	err = v.freeSpace(file.Info.Offset, file.Info.Size)
	if err != nil {
		return err
	}

	err = v.index.Delete(fid)
	return err
}

func (v *Volume)NewFile(size uint64) (*File, error) {
	v.rwMutex.Lock()
	defer v.rwMutex.Unlock()

	fid := v.newFid()
	offset, err := v.newSpace(size)
	if err != nil {
		return nil, err
	}

	fileInfo := &FileInfo{
		Fid: fid,
		Offset: offset,
		Size: size,
		Ctime: time.Now(),
		Mtime: time.Now(),
		Atime: time.Now(),
	}
	return &File{volume: v, Info: fileInfo}, nil
}

//TODO: newSpace
func (v *Volume)newSpace(size uint64) (offset uint64, err error) {
	//此函数不是并发安全
	return 0, nil
}

//TODO: freeSpace
func (v *Volume)freeSpace(offset uint64, size uint64) (err error) {
	//此函数不是并发安全
	return nil
}

func (v *Volume)truncate() {
	fi, err := v.dataFile.Stat()
	if err != nil {
		panic(err)
	}

	err = v.dataFile.Truncate(fi.Size() + 4294967296)
	if err != nil {
		panic(err)
	}
}

func (v *Volume)newFid() (fid uint64) {
	//此函数不是并发安全
	configKey := []byte("freeFid")
	b, err := v.index.getConfig(configKey)
	if err != nil || b == nil {
		fid = 0
		b := make([]byte, 8)
		b[0] = 1
		if err := v.index.setConfig(configKey, b); err != nil {
			panic(err)
		}
		return
	}else {
		fid = binary.LittleEndian.Uint64(b)
		binary.LittleEndian.PutUint64(b, fid + 1)
		if err := v.index.setConfig(configKey, b); err != nil {
			panic(err)
		}
		return
	}
}
