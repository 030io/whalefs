package api

import (
	"fmt"
	"os"
	"bytes"
	"mime/multipart"
	"path/filepath"
	"io"
	"net/http"
	"errors"
	"io/ioutil"
)

func Upload(host string, port int, dst string, src string) (err error) {
	fi, err := os.Stat(src)
	if os.IsNotExist(err) {
		return err
	}
	if fi.IsDir() {
		return fmt.Errorf("can't upload a directry: %s", src)
	}

	if dst[len(dst) - 1] == '/' {
		dst += filepath.Base(src)
	}

	var url string
	if dst[0] == '/' {
		url = fmt.Sprintf("http://%s:%d%s", host, port, dst)
	}else {
		url = fmt.Sprintf("http://%s:%d/%s", host, port, dst)
	}
	file, _ := os.Open(src)

	body := new(bytes.Buffer)
	mPart := multipart.NewWriter(body)

	filePart, err := mPart.CreateFormFile("file", filepath.Base(dst))
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
