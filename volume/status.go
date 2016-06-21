package volume

import (
	"github.com/syndtr/goleveldb/leveldb"
	"path/filepath"
	"strconv"
	"sync"
	"encoding/binary"
	"github.com/syndtr/goleveldb/leveldb/util"
	"errors"
)

var fidKey = []byte("\x00freeFid") //key="\x00freeFid" value=uint64(big endian 8byte)
var freeSpacePrefix = []byte("\x01FS") //key="\x02FS"+"offset(big endian 8byte)" value=size(big endian 8byte)
//var freeSpaceSizePrefix = []byte("\x02FSS") //"\x01FSS"+"size(big endian 8byte)"

type Status struct {
	path       string
	db         *leveldb.DB

	fidMutex   sync.Mutex
	spaceMutex sync.Mutex
}

func NewStatus(dir string, vid int) (status *Status, err error) {
	path := filepath.Join(dir, strconv.Itoa(vid) + ".status")
	status = new(Status)
	status.path = path
	status.db, err = leveldb.OpenFile(path, nil)
	return status, err
}

func (s *Status)newFid() (fid uint64) {
	s.fidMutex.Lock()
	defer s.fidMutex.Unlock()

	b, err := s.db.Get(fidKey, nil)
	if err != nil || b == nil {
		fid = 0
		b = make([]byte, 8)
	}else {
		fid = binary.BigEndian.Uint64(b)
	}

	binary.BigEndian.PutUint64(b, fid + 1)
	err = s.db.Put(fidKey, b, nil)
	if err != nil {
		panic(err)
	}

	return
}

func (s *Status)newSpace(size uint64) (offset uint64, err error) {
	s.spaceMutex.Lock()
	defer s.spaceMutex.Unlock()

	iter := s.db.NewIterator(util.BytesPrefix(freeSpacePrefix), nil)
	defer iter.Release()

	for iter.Next() {
		offset = binary.BigEndian.Uint64(iter.Key()[len(freeSpacePrefix):])
		freeSize := binary.BigEndian.Uint64(iter.Value())
		if freeSize < size {
			continue
		}

		transaction, err := s.db.OpenTransaction()
		if err != nil {
			return 0, err
		}

		transaction.Delete(iter.Key(), nil)

		if freeSize != size {
			key := make([]byte, 8)
			binary.BigEndian.PutUint64(key, offset + size)
			key = append(freeSpacePrefix, key...)

			value := make([]byte, 8)
			binary.BigEndian.PutUint64(value, freeSize - size)

			transaction.Put(key, value, nil)
		}

		err = transaction.Commit()
		if err != nil {
			return 0, err
		}

		return offset, nil
	}

	return 0, errors.New("can't new space: no free space")
}

func (s *Status)freeSpace(offset uint64, size uint64) error {
	s.spaceMutex.Lock()
	defer s.spaceMutex.Unlock()

	iter := s.db.NewIterator(util.BytesPrefix(freeSpacePrefix), nil)
	defer iter.Release()

	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, offset)
	key = append(freeSpacePrefix, key...)
	iter.Seek(key)

	transaction, err := s.db.OpenTransaction()
	if err != nil {
		return err
	}

	key = iter.Key()
	if len(key) == len(freeSpacePrefix) + 8 {
		nOffset := binary.BigEndian.Uint64(key[len(freeSpacePrefix):])
		if nOffset < offset + size {
			return errors.New("that is impossible")
		}else if nOffset == offset + size {
			transaction.Delete(key, nil)
			size += binary.BigEndian.Uint64(iter.Value())
		}
	}

	iter.Prev()
	key = iter.Key()
	if len(key) == len(freeSpacePrefix) + 8 {
		pOffset := binary.BigEndian.Uint64(key[len(freeSpacePrefix):])
		pSize := binary.BigEndian.Uint64(iter.Value())
		if pOffset + pSize > offset {
			return errors.New("that is impossible")
		}else if pOffset + pSize == offset {
			transaction.Delete(key, nil)
			offset = pOffset
			size += pSize
		}
	}

	key = make([]byte, 8)
	binary.BigEndian.PutUint64(key, offset)
	key = append(freeSpacePrefix, key...)

	value := make([]byte, 8)
	binary.BigEndian.PutUint64(value, size)
	transaction.Put(key, value, nil)

	return transaction.Commit()
}
