package volume

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
	"encoding/binary"
)

const truncateSize = 1024 * 1024 * 1024

type Volume struct {
	MaxSize  uint64

	index    Index
	status   *Status
	dataFile *os.File
	rwMutex  sync.RWMutex
}

func NewVolume(dir string, vid int) (v *Volume, err error) {
	path := filepath.Join(dir, strconv.Itoa(vid) + ".data")
	v = new(Volume)
	v.dataFile, err = os.OpenFile(path, os.O_CREATE | os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	v.index, err = NewLevelDBIndex(dir, vid)
	if err != nil {
		return nil, err
	}

	v.status, err = NewStatus(dir, vid)
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
		return &File{DataFile: v.dataFile, Info: fi}, nil
	}else {
		return nil, err
	}
}

func (v *Volume)Delete(fid uint64) error {
	file, err := v.Get(fid)
	if err != nil {
		return err
	}

	v.rwMutex.Lock()
	defer v.rwMutex.Unlock()

	//因为文件内容前后都写入fid(8 byte) 要一起释放
	err = v.status.freeSpace(file.Info.Offset - 8, file.Info.Size + 16)
	if err != nil {
		return err
	}

	err = v.index.Delete(fid)
	return err
}

func (v *Volume)NewFile(fileName string, size uint64) (f *File, err error) {
	v.rwMutex.Lock()
	defer v.rwMutex.Unlock()

	var fid uint64
	for {
		fid = v.status.newFid()
		if v.index.Has(fid) == false {
			break
		}
	}

	offset, err := v.newSpace(size + 16)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			if err := v.status.freeSpace(offset, size + 16); err != nil {
				panic(err)
			}
		}
	}()

	//在文件内容前后都写入fid
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, fid)
	_, err = v.dataFile.WriteAt(b, int64(offset))
	if err != nil {
		return nil, err
	}
	_, err = v.dataFile.WriteAt(b, int64(offset + 8 + size))
	if err != nil {
		return nil, err
	}

	fileInfo := &FileInfo{
		Fid: fid,
		Offset: offset + 8,
		Size: size,
		Ctime: time.Now(),
		Mtime: time.Now(),
		Atime: time.Now(),
		FileName: fileName,
	}

	err = v.index.Set(fileInfo)
	if err != nil {
		return nil, err
	}else {
		return &File{DataFile: v.dataFile, Info: fileInfo}, nil
	}
}

func (v *Volume)truncate() {
	fi, err := v.dataFile.Stat()
	if err != nil {
		panic(err)
	}

	err = v.dataFile.Truncate(fi.Size() + truncateSize)
	if err != nil {
		panic(err)
	}

	err = v.status.freeSpace(uint64(fi.Size()), truncateSize)
	if err != nil {
		panic(err)
	}
}

func (v *Volume)newSpace(size uint64) (uint64, error) {
	offset, err := v.status.newSpace(size)
	if err == nil {
		return offset, err
	}

	//TODO: check max size before truncate
	v.truncate()

	return v.status.newSpace(size)
}
