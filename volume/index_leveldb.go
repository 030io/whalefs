package volume

import (
	"github.com/syndtr/goleveldb/leveldb"
	"encoding/binary"
	"path/filepath"
	"strconv"
)

var configPrefix = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, '_'}

type LevelDBIndex struct {
	path string
	db   *leveldb.DB
}

func NewLevelDBIndex(dir string, vid int) (index *LevelDBIndex, err error) {
	path := filepath.Join(dir, strconv.Itoa(vid) + ".ldb")
	index = new(LevelDBIndex)
	index.path = path
	index.db, err = leveldb.OpenFile(path, nil)
	return index, err
}

func (l *LevelDBIndex)Get(fid uint64) (*FileInfo, error) {
	key := make([]byte, 8)
	binary.LittleEndian.PutUint64(key, fid)
	data, err := l.db.Get(key, nil)
	if err != nil {
		return nil, err
	}
	index := new(FileInfo)
	return index, index.UnMarshalBinary(data)
}

func (l *LevelDBIndex)Set(iv *FileInfo) error {
	data := iv.MarshalBinary()
	return l.db.Put(data[:8], data, nil)
}

func (l *LevelDBIndex)Delete(fid uint64) error {
	key := make([]byte, 8)
	binary.LittleEndian.PutUint64(key, fid)
	return l.db.Delete(key, nil)
}

func (l *LevelDBIndex)getConfig(key []byte) ([]byte, error) {
	return l.db.Get(append(configPrefix, key...), nil)
}

func (l *LevelDBIndex)setConfig(key, value []byte) error {
	return l.db.Put(append(configPrefix, key...), value, nil)
}
