package master

import (
	"gopkg.in/redis.v2"
	"fmt"
	"regexp"
	"strconv"
)

var urlReg *regexp.Regexp

func init() {
	var err error
	urlReg, err = regexp.Compile("/([0-9]*)/([0-9]*)/(.*)")
	if err != nil {
		panic(err)
	}
}

type MetadataRedis struct {
	client *redis.Client
}

func NewMetadataRedis(host string, port int, password string, database int) (*MetadataRedis, error) {
	mr := new(MetadataRedis)
	mr.client = redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr: fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB: int64(database),
	})

	if _, err := mr.client.Get("__key").Result(); err != nil && err != redis.Nil {
		return nil, err
	}

	return mr, nil
}

func (m *MetadataRedis)Get(filePath string) (vid uint64, fid uint64, fileName string, err error) {
	url, err := m.client.Get(filePath).Result()
	if err != nil {
		return
	}
	match := urlReg.FindStringSubmatch(url)
	vid, _ = strconv.ParseUint(match[1], 10, 64)
	fid, _ = strconv.ParseUint(match[2], 10, 64)
	fileName = match[3]
	return
}

func (m *MetadataRedis)Set(filePath string, vid uint64, fid uint64, fileName string) error {
	url := fmt.Sprintf("/%d/%d/%s", vid, fid, fileName)
	_, err := m.client.Set(filePath, url).Result()
	return err

}

func (m *MetadataRedis)Delete(filePath string) error {
	_, err := m.client.Del(filePath).Result()
	return err
}

func (m *MetadataRedis)Has(filePath string) bool {
	_, err := m.client.Get(filePath).Result()
	return err != redis.Nil
}

func (m *MetadataRedis)Close() error {
	return m.client.Close()
}
