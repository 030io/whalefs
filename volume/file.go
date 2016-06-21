package volume

import (
	"time"
	"encoding/binary"
)

type FileInfo struct {
	Fid      uint64
	Offset   uint64
	Size     uint64
	Ctime    time.Time
	Mtime    time.Time
	Atime    time.Time
	FileName string
}

func (iv *FileInfo)MarshalBinary() []byte {
	data := make([]byte, 48 + len(iv.FileName))
	binary.LittleEndian.PutUint64(data[0:8], iv.Fid)
	binary.LittleEndian.PutUint64(data[8:16], iv.Offset)
	binary.LittleEndian.PutUint64(data[16:24], iv.Size)
	binary.LittleEndian.PutUint64(data[24:32], uint64(iv.Ctime.Unix()))
	binary.LittleEndian.PutUint64(data[32:40], uint64(iv.Mtime.Unix()))
	binary.LittleEndian.PutUint64(data[40:48], uint64(iv.Atime.Unix()))
	copy(data[48:], []byte(iv.FileName))
	return data
}

func (iv *FileInfo)UnMarshalBinary(data []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	iv.Fid = binary.LittleEndian.Uint64(data[0:8])
	iv.Offset = binary.LittleEndian.Uint64(data[8:16])
	iv.Size = binary.LittleEndian.Uint64(data[16:24])
	iv.Ctime = time.Unix(int64(binary.LittleEndian.Uint64(data[24:32])), 0)
	iv.Mtime = time.Unix(int64(binary.LittleEndian.Uint64(data[32:40])), 0)
	iv.Atime = time.Unix(int64(binary.LittleEndian.Uint64(data[40:48])), 0)
	iv.FileName = string(data[48:])
	return err
}

type File struct {
	volume *Volume
	Info   *FileInfo
}

//TODO: read write
func (f *File)Read(b []byte) (n int, err error) {
	return 0, nil
}

func (f *File)Write(b []byte) (n int, err error) {
	return 0, nil
}
