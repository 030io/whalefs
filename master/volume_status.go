package master

import (
	"fmt"
	"net/http"
	"net/url"
	"io/ioutil"
	"github.com/030io/whalefs/manager/api"
	"bytes"
	"mime/multipart"
	"errors"
)

type VolumeStatus struct {
	Id           uint64
	DataFileSize uint64
	MaxFreeSpace uint64

	Writable     bool

	vmStatus     *VolumeManagerStatus `json:"-"`
}

func (vs *VolumeStatus)getFileUrl(fid uint64, fileName string) string {
	return fmt.Sprintf("http://%s:%d/%d/%d/%s", vs.vmStatus.PublicHost, vs.vmStatus.PublicPort, vs.Id, fid, fileName)
}

func (vs *VolumeStatus)redirectToFile(w http.ResponseWriter, r *http.Request, fid uint64, fileName string) {
	url := vs.getFileUrl(fid, fileName)
	http.Redirect(w, r, url, http.StatusFound)
}

func (vs *VolumeStatus)uploadFile(fid uint64, fileName string, data []byte) error {
	body := new(bytes.Buffer)
	mPart := multipart.NewWriter(body)

	filePart, err := mPart.CreateFormFile("file", fileName)
	if err != nil {
		return err
	}

	_, err = filePart.Write(data)
	if err != nil {
		return err
	}

	mPart.Close()

	url_, _ := url.Parse(
		fmt.Sprintf(
			"http://%s:%d/%d/%d/%s",
			vs.vmStatus.AdminHost,
			vs.vmStatus.AdminPort,
			vs.Id,
			fid,
			fileName))

	req, err := http.NewRequest(http.MethodPost, url_.String(), body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", mPart.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(fmt.Sprintf("%d != http.StatusCreated  body: %s", resp.StatusCode, body))
	}
	return nil
}

func (vs *VolumeStatus)delete(fid uint64, fileName string) error {
	return api.Delete(vs.vmStatus.AdminHost, vs.vmStatus.AdminPort, vs.Id, fid, fileName)
}

func (vs *VolumeStatus)isWritable(size uint64) bool {
	return vs.Writable && vs.MaxFreeSpace != 0 && vs.MaxFreeSpace >= size
}

func (vs *VolumeStatus)hasEnoughSpace() bool {
	return float64(vs.MaxFreeSpace) / float64(vs.DataFileSize) > 0.99
}

func volumesIsWritable(vStatusList []*VolumeStatus, size uint64) bool {
	for _, vs := range vStatusList {
		if !vs.isWritable(size) {
			return false
		}
	}
	return len(vStatusList) != 0
}

func volumesHasEnoughSpace(vStatusList []*VolumeStatus) bool {
	for _, vs := range vStatusList {
		if !vs.hasEnoughSpace() {
			return false
		}
	}
	return true
}
