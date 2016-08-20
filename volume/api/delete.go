package api

import (
	"net/http"
	"fmt"
	"io/ioutil"
	"errors"
)

func Delete(host string, port int, vid uint64, fid uint64, filename string) error {
	url := fmt.Sprintf("http://%s:%d/%d/%d/%s", host, port, vid, fid, filename)
	req, _ := http.NewRequest(http.MethodDelete, url, nil)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusAccepted || resp.StatusCode == http.StatusNotFound {
		return nil
	}else {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(string(body))
	}
}
