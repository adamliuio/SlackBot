// https://www.reddit.com/r/popular/new.json?sort=new

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/disintegration/imaging"
)

var RedditBearerToken RedditToken

const redditFilename string = "ids/ids-reddit.json"
const RedditOAuthUrl string = "https://oauth.reddit.com/"

const RedditTokenRetrivingUrl string = "https://www.reddit.com/api/v1/access_token"

type RedditToken struct {
	Access_token string `json:"access_token,omitempty"`
	Token_type   string `json:"token_type,omitempty"`
	Expires_in   int64  `json:"expires_in,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

type RedditRetrieve struct {
	Kind string             `json:"kind,omitempty"`
	Data RedditRetrieveData `json:"data,omitempty"`
}

type RedditRetrieveData struct {
	After    string                `json:"after,omitempty"`
	Children []RedditRetrieveChild `json:"children,omitempty"`
}

type RedditRetrieveChild struct {
	Kind string                  `json:"kind,omitempty"`
	Data RedditRetrieveChildData `json:"data,omitempty"`
}

type RedditRetrieveChildData struct {
	Author                 string `json:"author,omitempty"`
	Title                  string `json:"title,omitempty"`
	Thumbnail              string `json:"thumbnail,omitempty"`
	Url_overridden_by_dest string `json:"url_overridden_by_dest,omitempty"`
	Domain                 string `json:"domain,omitempty"`
	Is_reddit_media_domain string `json:"is_reddit_media_domain,omitempty"`
	Ups                    int    `json:"ups,omitempty"`
	Awards                 int    `json:"total_awards_received,omitempty"`
	Id                     string `json:"id,omitempty"`
	Post_hint              string `json:"post_hint,omitempty"` // data with "Selftext" have no "post_hint"
	Selftext               string `json:"selftext,omitempty"`
	Permalink              string `json:"permalink,omitempty"`
}

type RedditClient struct {
	CurrentSubReddit string
	savedIDs         []string
}

func (rc RedditClient) RetrieveNew(subReddit string) (err error) {
	var respBody []byte
	if respBody, err = rc.RetrieveList(subReddit); err != nil {
		return
	}
	var redditRetrieve RedditRetrieve
	if err = json.Unmarshal(respBody, &redditRetrieve); err != nil {
		return
	}
	rc.CurrentSubReddit = subReddit
	if err = rc.SendToSlack(redditRetrieve); err != nil {
		return
	}
	return
}

func (rc RedditClient) AutoRetrieveNew() (err error) {
	json.Unmarshal(utils.ReadFile(twitterFilename), &rc.savedIDs)
	var subReddits []string = strings.Split(os.Getenv("AutoRedditSubs"), ",")
	for _, subReddit := range subReddits {
		if err = rc.RetrieveNew(subReddit); err != nil {
			return
		}
	}
	j, _ := json.Marshal(rc.savedIDs)
	utils.WriteFile(j, redditFilename)
	return
}

func (rc RedditClient) SendToSlack(redditRetrieve RedditRetrieve) (err error) {
	// sort "storiesItemsList" base on scores
	sort.Slice(redditRetrieve.Data.Children, func(i, j int) bool {
		return redditRetrieve.Data.Children[i].Data.Ups > redditRetrieve.Data.Children[j].Data.Ups
	})

	var leastUps int
	leastUps, _ = strconv.Atoi(os.Getenv("AutoRedditLeaseScore"))
	var wg = sync.WaitGroup{}
	for _, kid := range redditRetrieve.Data.Children {
		var post RedditRetrieveChildData = kid.Data
		if post.Ups < leastUps {
			continue
		} else {
			rc.savedIDs = append(rc.savedIDs, post.Id)
		}
		wg.Add(1)
		go func(post RedditRetrieveChildData) {
			defer wg.Done()
			var mbs MessageBlocks
			var mbarr = []MessageBlock{}
			mbarr = append(mbarr, MessageBlock{Type: "divider"}, MessageBlock{
				Type: "context", Elements: []*Element{{Type: "plain_text", Text: rc.CurrentSubReddit, Emoji: true}},
			})
			if post.Post_hint == "" && post.Selftext != "" { // "text"
				mbs.Blocks = append(mbarr, rc.formatTextBlock(post)...)
			} else if post.Domain == "i.redd.it" { // "image"
				mbs.Blocks = append(mbarr, rc.formatImageBlock(post)...)
			} else if post.Domain == "v.redd.it" { // "video"
				mbs.Blocks = append(mbarr, rc.formatVideoBlock(post)...)
			} else { // "link"
				mbs.Blocks = append(mbarr, rc.formatLinkBlock(post)...)
			}
			if err = sc.SendBlocks(mbs, os.Getenv("SlackWebHookUrlReddit")); err != nil {
				return
			}
		}(post)
	}
	wg.Wait()
	return
}

func (rc RedditClient) formatLinkBlock(post RedditRetrieveChildData) (mbarr []MessageBlock) {
	mbarr = append(mbarr, sc.CreateTextBlock(fmt.Sprintf(`<%s|%s>`, post.Url_overridden_by_dest, post.Title), "mrkdwn", ""))
	mbarr = append(mbarr, sc.CreateTextBlock(fmt.Sprintf(`[<https://reddit.com%s|reddit>] Likes: *%d*`, post.Permalink, post.Ups), "mrkdwn", ""))
	return
}

func (rc RedditClient) formatVideoBlock(post RedditRetrieveChildData) (mbarr []MessageBlock) {
	mbarr = append(mbarr, sc.CreateTextBlock("*"+post.Title+"*", "mrkdwn", post.Thumbnail))
	mbarr = append(mbarr, sc.CreateTextBlock(fmt.Sprintf(`[<https://reddit.com%s|video>] Likes: *%d*`, post.Permalink, post.Ups), "mrkdwn", ""))
	return
}

func (rc RedditClient) formatTextBlock(post RedditRetrieveChildData) (mbarr []MessageBlock) {
	mbarr = append(mbarr, sc.CreateTextBlock("*"+post.Title+"*", "mrkdwn", ""))
	var selftext string
	if len(post.Selftext) > 1000 {
		selftext = post.Selftext[:1000] + " ..."
	} else {
		selftext = post.Selftext
	}
	mbarr = append(mbarr, sc.CreateTextBlock(selftext, "plain_text", ""))
	mbarr = append(mbarr, sc.CreateTextBlock(fmt.Sprintf(`[<https://reddit.com%s|reddit>] Likes: *%d*`, post.Permalink, post.Ups), "mrkdwn", ""))
	return
}

func (rc RedditClient) formatImageBlock(post RedditRetrieveChildData) (mbarr []MessageBlock) {
	mbarr = append(mbarr, sc.CreateTextBlock("*"+post.Title+"*", "mrkdwn", ""))
	var imageUrl string = post.Url_overridden_by_dest
	var imageSize int64
	var err error
	if _, imageSize, err = utils.CheckUrl(imageUrl); err != nil {
		return
	}
	if imageSize > int64(2000000) {
		var reg *regexp.Regexp = regexp.MustCompile(`\/([A-Za-z0-9])\w+.(jpg|png)`)
		var tempFolder string = "/tmp"
		var urlPrefix string = "https://naughtymonsta.digital/file"
		if Hostname == "MacBook-Pro.local" {
			tempFolder = "data-samples"
			urlPrefix = "https://i.redd.it"
		}
		var filePath string = tempFolder + reg.FindAllString(imageUrl, 1)[0]
		utils.DownloadFile(imageUrl, filePath, false)
		rc.ResizeImage(filePath)
		imageUrl = strings.ReplaceAll(filePath, tempFolder, "")
		mbarr = append(mbarr, sc.CreateImageBlock(urlPrefix+imageUrl, post.Title))
		mbarr = append(mbarr, sc.CreateTextBlock(fmt.Sprintf(`[<https://reddit.com%s|reddit>] Likes: *%d*, [<%s|og image>]`, post.Permalink, post.Ups, post.Url_overridden_by_dest), "mrkdwn", ""))
	} else {
		mbarr = append(mbarr, sc.CreateImageBlock(post.Url_overridden_by_dest, post.Title))
		mbarr = append(mbarr, sc.CreateTextBlock(fmt.Sprintf(`[<https://reddit.com%s|image>] Likes: *%d*`, post.Permalink, post.Ups), "mrkdwn", ""))
	}
	return
}

func (rc RedditClient) ResizeImage(filePath string) {
	var err error
	var srcImage image.Image

	if srcImage, err = imaging.Open(filePath); err != nil {
		log.Fatalf("failed to open image: %v", err)
	}
	var dstImage800 image.Image = imaging.Resize(srcImage, 800, 0, imaging.Lanczos)

	if err = imaging.Save(dstImage800, filePath); err != nil {
		log.Fatalf("failed to save image: %v", err)
	}
}

func (rc RedditClient) RetrievePost(postID string) (respBody []byte, err error) {
	var url string = "https://oauth.reddit.com/api/info/?id=t3_" + postID
	if respBody, err = utils.HttpRequest("GET", nil, url, [][]string{
		{"User-Agent", os.Getenv("AutoRedditLeaseScore")},
		{"Authorization", "bearer " + RedditBearerToken.Access_token},
	}); err != nil {
		return
	}
	return
}
func (rc RedditClient) RetrieveList(subReddit string) (respBody []byte, err error) {
	subReddit = strings.Trim(subReddit, "/")
	if time.Now().Unix() > time.Now().Unix()+RedditBearerToken.Expires_in {
		rc.RenewBearerToken()
	}
	var url string = RedditOAuthUrl + subReddit + "/hot"
	if respBody, err = utils.HttpRequest("GET", nil, url, [][]string{
		{"User-Agent", os.Getenv("AutoRedditLeaseScore")},
		{"Authorization", "bearer " + RedditBearerToken.Access_token},
	}); err != nil {
		// {"message": "Unauthorized", "error": 401}
		return
	}
	return
}

func (rc RedditClient) RenewBearerToken() (token RedditToken, err error) {
	params := url.Values{}
	params.Add("grant_type", `password`)
	params.Add("username", os.Getenv("RedditMyUsername"))
	params.Add("password", os.Getenv("RedditMyPassword"))
	var body *strings.Reader = strings.NewReader(params.Encode())

	var req *http.Request
	var resp *http.Response
	req, err = http.NewRequest("POST", RedditTokenRetrivingUrl, body)
	if err != nil {
		return
	}
	req.SetBasicAuth(os.Getenv("RedditAppID"), os.Getenv("RedditAppSecret"))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("User-Agent", os.Getenv("RedditRequestUserAgent"))

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var buf *bytes.Buffer = new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return
	}

	var respBody []byte = buf.Bytes()
	if err = json.Unmarshal(respBody, &token); err != nil {
		return
	}
	return
}
