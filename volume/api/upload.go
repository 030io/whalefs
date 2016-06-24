package api

import (
	"os"
	"bytes"
	"mime/multipart"
	"path/filepath"
	"io"
	"net/http"
	"errors"
	"fmt"
	"io/ioutil"
//"github.com/030io/whalefs/volume/manager"
//"encoding/json"
)

func Upload(host string, port int, vid int, fid uint64, filePath string, fileName string) (err error) {
	url := fmt.Sprintf("http://%s:%d/%d/%d/%s", host, port, vid, fid, fileName)
	file, err := os.Open(filePath)
	if os.IsNotExist(err) {
		return
	}

	body := new(bytes.Buffer)
	mPart := multipart.NewWriter(body)

	if fileName == "" {
		fileName = filepath.Base(filePath)
	}
	filePart, err := mPart.CreateFormFile("file", fileName)
	if err != nil {
		return
	}

	_, err = io.Copy(filePart, file)
	if err != nil {
		return
	}

	mPart.Close()

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", mPart.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		err = errors.New(fmt.Sprintf("%d != http.StatusCreated  body: %s", resp.StatusCode, body))
	}
	return
}
