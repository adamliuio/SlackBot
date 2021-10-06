package main

import (
	"encoding/json"
	"image"
	"log"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/disintegration/imaging"
)

func TestRedditAutoRetrieveNew(t *testing.T) {
	if err := rc.AutoRetrieveNew(); err != nil {
		log.Panicln(err)
	}
}

func TestGetSubRedditList(t *testing.T) {
	url := RedditOAuthUrl + "api/v1/me/prefs"
	rc.RenewBearerToken()

	respBody, err := utils.HttpRequest("GET", nil, url, [][]string{
		{"User-Agent", os.Getenv("AutoRedditLeaseScore")},
		{"Authorization", "bearer " + rc.RedditBearerToken.Access_token},
	})
	if err != nil {
		// {"message": "Unauthorized", "error": 401}
		return
	}
	t.Log(string(respBody))
}

func TestTime(t *testing.T) {
	t.Log(time.Now().Unix())
	time.Sleep(3 * time.Second)
	t.Log(time.Now().Unix())
}

func TestResizeImage(t *testing.T) {
	var url string = "https://i.redd.it/6nnx8o7aqbr71.jpg" //   3,663,667
	//  var url string="https://i.redd.it/oglf0dboo9r71.png" //  19,308,260

	var err error
	_, contentLength, err := utils.CheckUrl(url)
	if err != nil {
		return
	}
	t.Log(contentLength)
	if contentLength > int64(2000000) {
		var reg *regexp.Regexp = regexp.MustCompile(`\/([A-Za-z0-9])\w+.(jpg|png)`)
		var fn string = reg.FindAllString(url, 1)[0]
		utils.DownloadFile(url, "data-samples"+fn, false)
	}

	var fn string = "data-samples/6nnx8o7aqbr71.jpg"
	var srcImage image.Image

	if srcImage, err = imaging.Open(fn); err != nil {
		log.Fatalf("failed to open image: %v", err)
	}
	var dstImage800 image.Image = imaging.Resize(srcImage, 800, 0, imaging.Lanczos)
	err = imaging.Save(dstImage800, "data-samples/6nnx8o7aqbr71-800.jpg")
	if err != nil {
		log.Fatalf("failed to save image: %v", err)
	}
}

func TestFormatData(t *testing.T) {
	var err error
	var redditRetrieve RedditRetrieve
	_ = json.Unmarshal(utils.ReadFile("data-samples/reddit-space-hot.json"), &redditRetrieve)
	rc.CurrentSubReddit = "r/space"
	if err = rc.SendToSlack(redditRetrieve); err != nil {
		log.Panic(err)
	}
}

func TestRetrieveRedditPost(t *testing.T) {
	rc.RedditBearerToken = RedditToken{
		Access_token: "299518766060-7cxsDSSSOP7mV-KWFUQcBRVHYfirBg",
		Token_type:   "bearer",
		Expires_in:   3600,
		Scope:        "*",
	}

	var respBody []byte
	var err error
	if respBody, err = rc.RetrievePost("q0pf4t"); err != nil {
		t.Fatal(err)
	}
	t.Log(string(respBody))
}

func TestRedditRetrieve(t *testing.T) {
	// rc.RedditBearerToken = RedditToken{
	// 	Access_token: "299518766060-7cxsDSSSOP7mV-KWFUQcBRVHYfirBg",
	// 	Token_type:   "bearer",
	// 	Expires_in:   3600,
	// 	Scope:        "*",
	// }

	var respBody []byte
	var err error
	if respBody, err = rc.RetrieveList("r/Entrepreneur"); err != nil {
		t.Fatal(err)
	}
	// t.Log(string(respBody))
	utils.WriteFile(respBody, "data-samples/t.json")
}

func TestRedditGetBearerToken(t *testing.T) {
	var err error
	if rc.RedditBearerToken, err = rc.RenewBearerToken(); err != nil {
		t.Fatal(err)
	}

	t.Logf("bearerToken: %+v\n", rc.RedditBearerToken)
}
