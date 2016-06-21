package volume

import (
	"github.com/syndtr/goleveldb/leveldb"
	"encoding/binary"
)

type LevelDBIndex struct {
	fileName string
	db       *leveldb.DB
}

func NewLevelDBIndex(fileName string) (index *LevelDBIndex, err error) {
	index = new(LevelDBIndex)
	index.fileName = fileName
	index.db, err = leveldb.OpenFile(fileName, nil)
	return index, err
}

//TODO: get set delete
func (l *LevelDBIndex)Get(key uint64) (*IndexValue, error) {
	realKey := make([]byte, 8)
	binary.LittleEndian.PutUint64(realKey, key)
	data, err := l.db.Get(realKey, nil)
	if err != nil {
		return nil, err
	}
	index := new(IndexValue)
	return index, index.UnMarshalBinary(data)
}

func (l *LevelDBIndex)Set(iv *IndexValue) error {
	data := iv.MarshalBinary()
	return l.db.Put(data[:8], data, nil)
}

func (l *LevelDBIndex)Delete(key uint64) error {
	realKey := make([]byte, 8)
	binary.LittleEndian.PutUint64(realKey, key)
	return l.db.Delete(realKey, nil)
}
