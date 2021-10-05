// curl -X POST -H 'Content-type: application/json' --data '{"text":"Allow me to reintroduce myself!"}' YOUR_WEBHOOK_URL
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	urlUtils "net/url"
	"os"
	"strings"
	"testing"
)

func TestSendPlainText(t *testing.T) {
	var err error = sc.SendPlainText("what's *up* https://api.slack.com/reference/messaging/link-unfurling", os.Getenv("SlackWebHookUrlTest"))
	if err != nil {
		if strings.Contains(err.Error(), "Temporary failure in name resolution") {
			t.Log("err has Temporary failure in name resolution")
		}
		t.Fatal(err)
	}
}

func TestSendMarkdownText(t *testing.T) {
	tt := "📺 */command*: returns all your commands for you to see\n📰 */hn* (/hn top 10-20) returns a list of buttons for retrieving buttons to interact with Hacker News."
	var err error = sc.SendMarkdownText(tt, os.Getenv("SlackWebHookUrlTest"), "")
	if err != nil {
		t.Fatal(err)
	}
}

func TestUrlRedirect(t *testing.T) {
	t.Log(utils.CheckUrl("https://t.co/nxURNgmEvk"))
}

func TestPostToSlack(t *testing.T) {
	var err error
	var req *http.Request
	var resp *http.Response
	var jsonStr string = `{"text":"Allow me to reintroduce myself!"}`
	var reqBody = []byte(jsonStr)
	req, err = http.NewRequest(http.MethodPost, os.Getenv("SlackWebHookUrlTest"), bytes.NewBuffer(reqBody))
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(body))
}

func TestStringSplit(t *testing.T) {
	t.Logf("%q\n", strings.Split(os.Getenv("AutoRedditSub"), " "))
}

func TestDeleteMsg(t *testing.T) {
	var err error = sc.DeleteMsg("C02FLJCFSJC", "1633150962.002000")
	if err != nil {
		t.Fatal(err)
	}
}

func TestAuth(t *testing.T) {
	// curl -i -H "Authorization:Bearer xoxb-" -X POST https://slack.com/api/auth.test
	var url string = "https://slack.com/api/auth.test"
	var headers = [][]string{{"Authorization", fmt.Sprintf("Bearer %s", os.Getenv("SlackBotUserOAuthToken"))}}
	var err error = sc.SendBytes([]byte{}, url, headers)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUrl(t *testing.T) {
	u, err := urlUtils.Parse("https://siongui.github.io/pali-chanting/zh/archives.html")
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(u.Hostname(), ".")
	log.Printf("u.Hostname(): %+v\n\n", u.Hostname())
	log.Printf("parts: %+v\n\n", parts)
	domain := parts[len(parts)-2] + "." + parts[len(parts)-1]
	log.Println(domain)
}
