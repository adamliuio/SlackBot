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
	"runtime/debug"
	"strings"
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
		log.Panicf("%s not found: %s", filename, err)
	}
	return
}

func (u Utils) WriteFile(b []byte, filename string) (err error) {
	return ioutil.WriteFile(filename, b, 0644)
}

func (u Utils) DownloadFile(url, fn string, ignoreErr bool) {

	// Create blank file
	var file *os.File
	var err error
	if file, err = os.Create(fn); err != nil {
		u.DealWithError(err)
	}
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}

	// Put content on file
	var resp *http.Response
	if resp, err = client.Get(url); err != nil {
		u.DealWithError(err)
	}
	defer resp.Body.Close()

	if _, err = io.Copy(file, resp.Body); err != nil {
		u.DealWithError(err)
	}
	defer file.Close()
}

func (u Utils) DealWithError(err error) {
	var now, errMsg, wd string
	var e error
	if wd, e = os.Getwd(); e != nil {
		log.Panic(e)
	}
	now, _ = utils.TimeNow()
	errMsg = fmt.Sprintf("%s\nerr: %s\ndetail: %s", now, err.Error(), strings.Replace(string(debug.Stack()), wd, ".", -1))
	if IsTestMode && IsLocal {
		log.Fatalln(errMsg)
	} else {
		sc.SendPlainText(errMsg, os.Getenv("SlackWebHookUrlTest"))
	}
}

func (u Utils) ConvertUnixTime(unixTs int64) (tm string) {
	tm = time.Unix(unixTs, 0).Format("2006-01-02")
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

func (u Utils) TimeNow() (string, int64) {
	var loc *time.Location
	var err error
	if loc, err = time.LoadLocation(Params.Timezone); err != nil {
		panic(err)
	}
	var now time.Time = time.Now().In(loc)
	return now.Format(Params.TimeFormat), now.Unix()
}
