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
	return mr, nil
}

func (m *MetadataRedis)Get(filePath string) (vid int, fid uint64, fileName string, err error) {
	url, err := m.client.Get(filePath).Result()
	if err != nil {
		return
	}
	match := urlReg.FindStringSubmatch(url)
	vid, _ = strconv.Atoi(match[1])
	fid, _ = strconv.ParseUint(match[2], 10, 64)
	fileName = match[3]
	return
}

func (m *MetadataRedis)Set(filePath string, vid int, fid uint64, fileName string) error {
	url := fmt.Sprintf("/%d/%d/%s", vid, fid, fileName)
	_, err := m.client.Set(filePath, url).Result()
	return err

}

func (m *MetadataRedis)Delete(filePath string) error {
	_, err := m.client.Del(filePath).Result()
	return err
}

func (m *MetadataRedis)setConfig(key string, value string) error {
	_, err := m.client.Set(key, value).Result()
	return err
}

func (m *MetadataRedis)getConfig(key string) (string, error) {
	return m.client.Get(key).Result()
}
