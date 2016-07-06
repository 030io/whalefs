package api

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"errors"
)

func Delete(host string, port int, filePath string) error {
	if filePath[0] == '/' {
		filePath = filePath[1:]
	}
	var url string
	if filePath[0] == '/' {
		url = fmt.Sprintf("http://%s:%d%s", host, port, filePath)
	}else {
		url = fmt.Sprintf("http://%s:%d/%s", host, port, filePath)
	}
	req, _ := http.NewRequest(http.MethodDelete, url, nil)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusAccepted {
		return nil
	}else {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(string(body))
	}
}
