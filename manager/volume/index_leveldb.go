package volume

import (
	"github.com/syndtr/goleveldb/leveldb"
	"encoding/binary"
	"path/filepath"
	"strconv"
)

type LevelDBIndex struct {
	path string
	db   *leveldb.DB
}

func NewLevelDBIndex(dir string, vid uint64) (index *LevelDBIndex, err error) {
	path := filepath.Join(dir, strconv.FormatUint(vid, 10) + ".index")
	index = new(LevelDBIndex)
	index.path = path
	index.db, err = leveldb.OpenFile(path, nil)
	return index, err
}

func (l *LevelDBIndex)Has(fid uint64) bool {
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, fid)
	_, err := l.db.Get(key, nil)
	return err == nil
}

func (l *LevelDBIndex)Get(fid uint64) (*FileInfo, error) {
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, fid)
	data, err := l.db.Get(key, nil)
	if err != nil {
		return nil, err
	}
	fi := new(FileInfo)
	return fi, fi.UnMarshalBinary(data)
}

func (l *LevelDBIndex)Set(fi *FileInfo) error {
	data := fi.MarshalBinary()
	return l.db.Put(data[:8], data, nil)
}

func (l *LevelDBIndex)Delete(fid uint64) error {
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, fid)
	return l.db.Delete(key, nil)
}

func (l *LevelDBIndex)Close() error {
	return l.db.Close()
}
