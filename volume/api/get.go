package api

import (
	"net/http"
	"fmt"
	"io/ioutil"
)

const bufferSize = 512 * 1024

func Get(host string, port int, vid uint64, fid uint64, filename string) ([]byte, error) {
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

func GetRange(host string, port int, vid uint64, fid uint64, filename string, start int, length int) ([]byte, error) {
	url := fmt.Sprintf("http://%s:%d/%d/%d/%s", host, port, vid, fid, filename)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, start + length - 1))
	resp, err := http.DefaultClient.Do(req)
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
