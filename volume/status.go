package volume

import (
	"github.com/syndtr/goleveldb/leveldb"
	"sync"
	"encoding/binary"
	"github.com/syndtr/goleveldb/leveldb/util"
	"errors"
	"path/filepath"
	"strconv"
	"fmt"
)

var reversedsizeOffset []byte //key= "\x01"+Reversesize(8 byte)+offset(8 byte) value=[]
var offsetSize []byte //key= "\x02"+offset(8 byte)+size(8 byte) value=[]

func init() {
	reversedsizeOffset = make([]byte, 1 + 16)
	offsetSize = make([]byte, 1 + 16)
	reversedsizeOffset[0] = '\x11'
	offsetSize[0] = '\x22'
}

type Status struct {
	path       string
	db         *leveldb.DB

	spaceMutex sync.Mutex
}

func NewStatus(dir string, vid uint64) (status *Status, err error) {
	path := filepath.Join(dir, strconv.FormatUint(vid, 10) + ".status")
	status = new(Status)
	status.path = path
	status.db, err = leveldb.OpenFile(path, nil)
	return status, err
}

func (s *Status)newSpace(size uint64) (offset uint64, err error) {
	s.spaceMutex.Lock()
	defer s.spaceMutex.Unlock()

	//这里根据size倒序存储,使最大的空间最先被获取
	iter := s.db.NewIterator(util.BytesPrefix(reversedsizeOffset[:1]), nil)
	defer iter.Release()

	iter.Next()
	key := iter.Key()
	if len(key) == 0 {
		return 0, errors.New("can't get free space")
	}

	freeSize := binary.BigEndian.Uint64(key[1:9]) ^ (^uint64(0))
	if freeSize < size {
		return 0, errors.New("can't get free space")
	}
	offset = binary.BigEndian.Uint64(key[9:])

	transaction, err := s.db.OpenTransaction()
	if err != nil {
		return 0, err
	}

	transaction.Delete(key, nil)

	key = getOffsetSizeKey(offset, freeSize)
	transaction.Delete(key, nil)

	if freeSize > size {
		key = getReversedsizeOffset(offset + size, freeSize - size)
		transaction.Put(key, nil, nil)

		key = getOffsetSizeKey(offset + size, freeSize - size)
		transaction.Put(key, nil, nil)
	}

	err = transaction.Commit()
	return offset, err
}

func (s *Status)freeSpace(offset uint64, size uint64) error {
	s.spaceMutex.Lock()
	defer s.spaceMutex.Unlock()

	iter := s.db.NewIterator(util.BytesPrefix(offsetSize[:1]), nil)
	defer iter.Release()

	key := getOffsetSizeKey(offset, 0)
	iter.Seek(key)

	transaction, err := s.db.OpenTransaction()
	if err != nil {
		return err
	}

	key = iter.Key()
	if len(key) != 0 {
		nOffset := binary.BigEndian.Uint64(key[1:9])
		nSize := binary.BigEndian.Uint64(key[9:])
		if nOffset < offset + size {
			panic(fmt.Errorf("nOffset: %d < offset: %d + size: %d", nOffset, offset, size))
			//if nOffset == offset {
			//	transaction.Discard()
			//return errors.New("space already free")
		} else if nOffset == offset + size {
			transaction.Delete(key, nil)
			size += nSize

			key = getReversedsizeOffset(nOffset, nSize)
			transaction.Delete(key, nil)
		}
	}

	iter.Prev()
	key = iter.Key()
	if len(key) != 0 {
		pOffset := binary.BigEndian.Uint64(key[1:9])
		pSize := binary.BigEndian.Uint64(key[9:])
		if pOffset + pSize > offset {
			panic(fmt.Errorf("pOffset: %d + pSize: %d > offset: %d", pOffset, pSize, offset))
			//transaction.Discard()
			//return errors.New("space alread free")
		} else if pOffset + pSize == offset {
			transaction.Delete(key, nil)
			offset = pOffset
			size += pSize

			key = getReversedsizeOffset(pOffset, pSize)
			transaction.Delete(key, nil)
		}
	}

	key = getOffsetSizeKey(offset, size)
	transaction.Put(key, nil, nil)

	key = getReversedsizeOffset(offset, size)
	transaction.Put(key, nil, nil)

	return transaction.Commit()
}

func getOffsetSizeKey(offset, size uint64) []byte {
	key := make([]byte, 1 + 16)
	key[0] = offsetSize[0]
	binary.BigEndian.PutUint64(key[1:9], offset)
	binary.BigEndian.PutUint64(key[9:], size)
	return key
}

func getReversedsizeOffset(offset, size uint64) []byte {
	key := make([]byte, 1 + 16)
	key[0] = reversedsizeOffset[0]
	binary.BigEndian.PutUint64(key[9:], offset)
	binary.BigEndian.PutUint64(key[1:9], size ^ (^uint64(0)))
	return key
}
