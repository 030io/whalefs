package api

import (
	"github.com/030io/whalefs/master"
	"fmt"
	"encoding/json"
	"bytes"
	"net/http"
	"io/ioutil"
)

func Heartbeat(host string, port int, vms *master.VolumeManagerStatus) error {
	url := fmt.Sprintf("http://%s:%d/heartbeat", host, port)
	body, err := json.Marshal(vms)
	reader := bytes.NewReader(body)
	resp, err := http.Post(url, "application/json", reader)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ = ioutil.ReadAll(resp.Body)
		return fmt.Errorf("%d != 200  body: %s", resp.StatusCode, body)
	}
	return nil
}
