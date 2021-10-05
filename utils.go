package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type Utils struct{}

func (u Utils) HttpRequest(requestMethod string, reqBody []byte, url string, headers [][]string) (respBody []byte, err error) {
	var req *http.Request
	var resp *http.Response
	req, err = http.NewRequest(requestMethod, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return
	}
	if len(headers) > 0 {
		for _, header := range headers {
			req.Header.Add(header[0], header[1])
		}
	}

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		return
	}

	var buf *bytes.Buffer = new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return
	}

	respBody = buf.Bytes()
	return
}

func (u Utils) ReadFile(filename string) (f []byte) {
	var err error
	if f, err = ioutil.ReadFile(filename); err != nil {
		log.Printf("%s not found", filename)
	}
	return
}

func (u Utils) WriteFile(b []byte, filename string) (err error) {
	return ioutil.WriteFile(filename, b, 0644)
}

func (u Utils) DownloadFile(url, fn string, ignoreErr bool) {

	// Create blank file
	file, err := os.Create(fn)
	u.dealWithError(err, fn, url, ignoreErr)
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	// Put content on file
	resp, err := client.Get(url)
	u.dealWithError(err, fn, url, ignoreErr)
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	u.dealWithError(err, fn, url, ignoreErr)
	defer file.Close()
}

func (u Utils) dealWithError(err error, fn, url string, ignoreErr bool) {
	if err != nil {
		if ignoreErr {
			log.Println(err)
			sc.SendPlainText(fmt.Sprintf(`Error: %s\nwhen downloading "%s"\nfrom "%s"`, err.Error(), fn, url), os.Getenv("SlackWebHookUrlTest"))
		} else {
			log.Panic(err)
		}
	}
}

func (u Utils) ConvertUnixTime(unixTs int) (tm string) {
	tm = time.Unix(int64(unixTs), 0).Format("01-02")
	return
}

func (u Utils) PrettyJsonString(body []byte) (respJson string) {
	dst := &bytes.Buffer{}
	if err := json.Indent(dst, body, "", "  "); err != nil {
		log.Panic(err)
	}
	respJson = dst.String()
	return
}

func (u Utils) CheckUrl(url string) (finalUrl string, contentLength int64, err error) {
	// check redirected final url & remove file size
	var resp *http.Response
	if resp, err = http.Head(url); err != nil {
		return
	}

	finalUrl = resp.Request.URL.String()
	contentLength = resp.ContentLength
	return
}
