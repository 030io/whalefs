package api

import (
	"net/http"
	"io/ioutil"
	"errors"
	"fmt"
)

func CreateVolume(host string, port int, vid uint64) error {
	url := fmt.Sprintf("http://%s:%d/%d/", host, port, vid)
	resp, err := http.Post(url, "text/plain", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		return nil
	}else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return errors.New(string(body))
	}
}
