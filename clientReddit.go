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
	"strings"
	"sync"
	"time"

	"github.com/disintegration/imaging"
)

type RedditToken struct {
	Access_token string `json:"access_token,omitempty"`
	Token_type   string `json:"token_type,omitempty"`
	Expires_in   int64  `json:"expires_in,omitempty"`
	Scope        string `json:"scope,omitempty"`
	Renewed_at   int64
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
	Is_reddit_media_domain bool   `json:"is_reddit_media_domain,omitempty"`
	Ups                    int    `json:"ups,omitempty"`
	Awards                 int    `json:"total_awards_received,omitempty"`
	Id                     string `json:"id,omitempty"`
	Post_hint              string `json:"post_hint,omitempty"` // data with "Selftext" have no "post_hint"
	Selftext               string `json:"selftext,omitempty"`
	Permalink              string `json:"permalink,omitempty"`
}

type RedditClient struct {
	CurrentSubReddit  string
	CurrentProfile    string
	savedIDs          []string
	newIdBatches      map[string][]string
	RedditBearerToken RedditToken
	RetrieveStats     map[string]int
}

func (rc RedditClient) AutoRetrieveNew() (err error) {
	rc.RetrieveStats = make(map[string]int)
	rc.newIdBatches = make(map[string][]string)
	json.Unmarshal(utils.ReadFile(redditFilename), &rc.savedIDs)
	for _, profile := range []string{"Adam", "Logen"} {
		rc.CurrentProfile = profile
		var subReddits []string = strings.Split(os.Getenv("AutoRedditSubs"+profile), ",")
		for _, subReddit := range subReddits {
			rc.CurrentSubReddit = subReddit
			if err = rc.RetrieveNew(subReddit); err != nil {
				return
			}
		}
	}
	var statStr string = "Reddit:"
	for sub, num := range rc.RetrieveStats {
		rc.savedIDs = append(rc.savedIDs, rc.newIdBatches[sub]...)
		statStr += fmt.Sprintf("\n%s: %d", sub, num)
	}
	j, _ := json.Marshal(rc.savedIDs)
	utils.WriteFile(j, redditFilename)
	log.Println(statStr)
	return
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
	if err = rc.SendToSlack(redditRetrieve); err != nil {
		return
	}
	return
}

func (rc RedditClient) SendToSlack(redditRetrieve RedditRetrieve) (err error) {
	// sort "storiesItemsList" base on scores
	sort.Slice(redditRetrieve.Data.Children, func(i, j int) bool {
		return redditRetrieve.Data.Children[i].Data.Ups > redditRetrieve.Data.Children[j].Data.Ups
	})

	var leastUps, qualifiedCount int
	leastUps = Params.AutoRedditLeaseScore
	var wg = sync.WaitGroup{}
	for _, kid := range redditRetrieve.Data.Children {
		qualifiedCount++
		var post RedditRetrieveChildData = kid.Data
		var exist bool = false
		for _, existId := range rc.savedIDs { // don't save saved ids
			if post.Id == existId {
				exist = true
				break
			}
		}
		if post.Ups < leastUps || exist {
			continue
		} else {
			rc.newIdBatches[rc.CurrentSubReddit] = append(rc.newIdBatches[rc.CurrentSubReddit], post.Id)
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
			if err = sc.SendBlocks(mbs, os.Getenv("SlackWebHookUrlReddit"+rc.CurrentProfile)); err != nil {
				return
			}
		}(post)
	}
	wg.Wait()
	rc.RetrieveStats[rc.CurrentSubReddit] = qualifiedCount
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
	if len(post.Selftext) > 500 {
		selftext = post.Selftext[:500] + " ..."
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
	if imageSize > int64(2000000) && Hostname != "MacBook-Pro.local" { // if image is bigger than 2mb & not on my local computer
		var reg *regexp.Regexp = regexp.MustCompile(`\/([A-Za-z0-9])\w+.(jpg|png)`)
		var tempFolder string = "/tmp"
		var urlPrefix string = "https://naughtymonsta.digital/file"
		if len(reg.FindAllString(imageUrl, 1)) > 0 {
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
		{"User-Agent", fmt.Sprintf("%d", Params.AutoRedditLeaseScore)},
		{"Authorization", "bearer " + rc.RedditBearerToken.Access_token},
	}); err != nil {
		return
	}
	return
}

func (rc RedditClient) RetrieveList(subReddit string) (respBody []byte, err error) {
	subReddit = strings.Trim(subReddit, "/")
	if time.Now().Unix() > rc.RedditBearerToken.Renewed_at+rc.RedditBearerToken.Expires_in-60 { // give it 60 seconds of space
		rc.RenewBearerToken()
	}
	var url string = RedditOAuthUrl + subReddit + "/hot.json?raw_json=1"
	if respBody, err = utils.HttpRequest("GET", nil, url, [][]string{
		{"User-Agent", fmt.Sprintf("%d", Params.AutoRedditLeaseScore)},
		{"Authorization", "bearer " + rc.RedditBearerToken.Access_token},
	}); err != nil {
		// {"message": "Unauthorized", "error": 401}
		return
	}
	return
}

func (rc RedditClient) RenewBearerToken() (token RedditToken, err error) {
	urlParams := url.Values{}
	urlParams.Add("grant_type", `password`)
	urlParams.Add("username", os.Getenv("RedditMyUsername"))
	urlParams.Add("password", os.Getenv("RedditMyPassword"))
	var body *strings.Reader = strings.NewReader(urlParams.Encode())

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
	token.Renewed_at = time.Now().Unix()
	return
}
