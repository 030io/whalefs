package api

import (
	"net/http"
	"fmt"
	"io/ioutil"
)

const bufferSize = 512 * 1024

//返回的resp.body,需要被关闭
func GetFile(host string, port int, vid int, fid uint64, filename string) ([]byte, error) {
	url := fmt.Sprintf("http://%s:%d/%d/%d/%s", host, port, vid, fid, filename)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return ioutil.ReadAll(resp.Body)
	}else {
		return nil, fmt.Errorf("%d != 200", resp.StatusCode)
	}
}
