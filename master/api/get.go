package api

import (
	"net/http"
	"fmt"
	"io/ioutil"
)

func Get(host string, port int, filePath string) ([]byte, error) {
	if filePath[0] == '/' {
		filePath = filePath[1:]
	}
	url := fmt.Sprintf("http://%s:%d/%s", host, port, filePath)
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
