package volume

import (
	"os"
	"path/filepath"
	"strconv"
)

type Volume struct {
	Index    Index
	DataFile *os.File
}

func NewVolume(dir string, vid int) (v *Volume, err error) {
	path := filepath.Join(dir, strconv.Itoa(vid) + ".data")
	v.DataFile, err = os.OpenFile(path, os.O_CREATE | os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	v.Index, err = NewLevelDBIndex(dir, vid)
	if err != nil {
		return nil, err
	}

	return v, nil
}

//TODO: get add delete
func (v *Volume)Get(fid uint64) *File {
	return nil
}

func (v *Volume)Add(*File) error {
	return nil
}

func (v *Volume)Delete(fid uint64) error {
	return nil
}
