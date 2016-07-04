package volume

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
	"encoding/binary"
	"errors"
)

const truncateSize = 1024 * 1024 * 1024

type Volume struct {
	Id           int
	MaxSize      uint64
	DataFileSize uint64

	index        Index
	status       *Status
	dataFile     *os.File
	mutex        sync.Mutex
}

func NewVolume(dir string, vid int) (v *Volume, err error) {
	path := filepath.Join(dir, strconv.Itoa(vid) + ".data")
	v = new(Volume)
	v.Id = vid
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

	fi, _ := v.dataFile.Stat()
	v.DataFileSize = uint64(fi.Size())

	return v, nil
}

func (v *Volume)Get(fid uint64) (*File, error) {
	//v.rwMutex.RLock()
	//defer v.rwMutex.RUnlock()

	fi, err := v.index.Get(fid)
	if err == nil {
		return &File{DataFile: v.dataFile, Info: fi}, nil
	}else {
		return nil, err
	}
}

func (v *Volume)Delete(fid uint64, fileName string) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	file, err := v.Get(fid)
	if err != nil {
		return err
	}else if file.Info.FileName != fileName {
		return errors.New("filename is wrong")
	}

	//因为文件内容前后都写入fid(8 byte) 要一起释放
	err = v.status.freeSpace(file.Info.Offset - 8, file.Info.Size + 16)
	if err != nil {
		return err
	}

	err = v.index.Delete(fid)
	return err
}

func (v *Volume)NewFile(fid uint64, fileName string, size uint64) (f *File, err error) {
	//v.mutex.Lock()
	//defer v.mutex.Unlock()

	//var fid uint64
	//for {
	//	fid = v.status.newFid()
	//	if v.index.Has(fid) == false {
	//		break
	//	}
	//}
	if v.index.Has(fid) {
		return nil, errors.New("fid is existed")
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
	v.mutex.Lock()
	defer v.mutex.Unlock()

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

	v.DataFileSize = uint64(fi.Size() + truncateSize)
}

func (v *Volume)newSpace(size uint64) (uint64, error) {
	offset, err := v.status.newSpace(size)
	if err == nil {
		return offset, err
	}

	v.truncate()

	return v.status.newSpace(size)
}

func (v *Volume)Close() {
	v.mutex.Lock()
	//v.status.spaceMutex.Lock()
	//因为要退出,所以不解锁,禁止读写
	//defer v.rwMutex.Unlock()

	//将所有资源释放
	v.dataFile.Close()
	v.dataFile = nil

	v.status.db.Close()
	v.status = nil

	v.index.Close()
	v.index = nil
}
