package check

import (
	"os"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/030io/whalefs/manager/volume"
	"encoding/binary"
	log "github.com/Sirupsen/logrus"
	_ "github.com/030io/whalefs/utils/logrus_hook"
)

type Checker struct {
	dataFile *os.File
	index    *leveldb.DB
	status   *leveldb.DB
}

func NewChecker(volumePath string) (*Checker, error) {
	checker := new(Checker)

	var err error
	checker.dataFile, err = os.Open(volumePath)
	if err != nil {
		return checker, err
	}

	checker.index, err = leveldb.OpenFile(volumePath[:len(volumePath) - 5] + ".index", nil)
	if err != nil {
		return checker, err
	}

	checker.status, err = leveldb.OpenFile(volumePath[:len(volumePath) - 5] + ".status", nil)
	return checker, err
}

func (c *Checker)Check() {
	checkDB, err := leveldb.Open(storage.NewMemStorage(), nil)
	if err != nil {
		panic(err)
	}

	statusIter := c.status.NewIterator(util.BytesPrefix([]byte{volume.OffsetSizePrefix}), nil)
	defer statusIter.Release()

	for statusIter.Next() {
		key := statusIter.Key()
		if has, err := checkDB.Has(key[1:], nil); err == nil {
			if has {
				offset := binary.BigEndian.Uint64(key[1:9])
				size := binary.BigEndian.Uint64(key[9:])
				log.WithFields(log.Fields{
					"offset": offset,
					"size": size,
				}).Error("existed!")
			} else {
				err = checkDB.Put(key[1:], nil, nil)
				if err != nil {
					panic(err)
				}
			}
		}
	}

	indexIter := c.index.NewIterator(nil, nil)
	defer indexIter.Release()

	for indexIter.Next() {
		data := indexIter.Value()
		file := new(volume.FileInfo)
		err := file.UnMarshalBinary(data)
		if err == nil {
			key := make([]byte, 16)
			//前后各有一个8字节的fid, offset向前8字节, size增加16字节
			binary.BigEndian.PutUint64(key[:8], file.Offset - 8)
			binary.BigEndian.PutUint64(key[8:], file.Size + 16)
			err = checkDB.Put(key, nil, nil)
			if err != nil {
				panic(err)
			}
		} else {
			log.WithFields(log.Fields{"data": data, "err": err.Error()}).Error()
		}
	}

	cIter := checkDB.NewIterator(nil, nil)
	defer cIter.Release()

	var (
		offset uint64
		size uint64
		key = make([]byte, 16)
		nOffset uint64
		nSize uint64
		nKey = make([]byte, 16)
	)
	cIter.Next()
	key = cIter.Key()
	offset = binary.BigEndian.Uint64(key[:8])
	size = binary.BigEndian.Uint64(key[8:])
	for cIter.Next() {
		nKey = cIter.Key()
		nOffset = binary.BigEndian.Uint64(nKey[:8])
		nSize = binary.BigEndian.Uint64(nKey[8:])

		if offset + size != nOffset {
			log.WithFields(log.Fields{
				"offset": offset,
				"size": size,
				"nOffset": nOffset,
				"nSize": nSize,
			}).Error()
		}

		offset = nOffset
		size = nSize
	}

	datafileStat, err := c.dataFile.Stat()
	if err != nil {
		panic(err)
	}
	if offset + size != uint64(datafileStat.Size()) {
		log.WithFields(log.Fields{
			"offset": offset,
			"size": size,
			"datafileSize": datafileStat.Size(),
		}).Error()
	}
}

func (c *Checker)Close() {
	c.dataFile.Close()
	c.index.Close()
	c.status.Close()
}

func Check(path string) {
	fi, err := os.Stat(path)
	if err != nil {
		panic(err)
	}
	if fi.IsDir() {
		//TODO: check volume dir
	} else {
		checker, err := NewChecker(path)
		if err != nil {
			panic(err)
		}
		checker.Check()
	}
}
