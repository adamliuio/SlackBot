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

func (u Utils) GetItemById(formatStr string, id int) (item HNItem) {
	var url string = fmt.Sprintf(formatStr, id)
	var body []byte = u.RetrieveBytes(url)

	if err := json.Unmarshal(body, &item); err != nil {
		log.Fatalln(err)
	}
	return
}

func (u Utils) RetrieveBytes(url string) (body []byte) {
	var resp *http.Response
	var err error
	resp, err = http.Get(url)

	if err != nil {
		log.Panicln(err)
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panicln(err)
	}

	return
}

func (u Utils) SendBytes(reqBody []byte, url string) (err error) {
	var req *http.Request
	req, err = http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return
	}
	if buf.String() != "ok" && buf.String() != `{"ok":true}` {
		err = fmt.Errorf("non-ok response returned from Slack, message: %s", buf.String())
		return
	}
	return
}

func (u Utils) ReadFile(filename string) (f []byte) {
	var err error
	f, err = ioutil.ReadFile(filename)
	if err != nil {
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
			sc.SendPlainText(fmt.Sprintf(`Error: %s\nwhen downloading "%s"\nfrom "%s"`, err.Error(), fn, url), sc.WebHookUrlTest)
		} else {
			log.Panic(err)
		}
	}
}

func (u Utils) ConvertUnixTime(unixTs int) (tm string) {
	tm = time.Unix(int64(unixTs), 0).Format("01-02")
	return
}
