// https://www.reddit.com/r/popular/new.json?sort=new

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/turnage/graw/reddit"
)

const redditFilename string = "ids-reddit.json"

type RedditClient struct {
	WebHookUrlReddit   string
	retrivedStoriesIds []string
}

func (rc RedditClient) Init() {
	_ = json.Unmarshal(utils.ReadFile(redditFilename), &rc.retrivedStoriesIds)
}

func (rc RedditClient) Graw() (harvest reddit.Harvest) {
	var err error
	var bot reddit.Bot
	bot, err = reddit.NewBotFromAgentFile("reddit.agent", 0)
	if err != nil {
		log.Fatalln("Failed to create bot handle: ", err)
	}
	harvest, err = bot.Listing(os.Getenv("AutoRedditSub"), "")
	if err != nil {
		log.Fatalln("Failed to fetch /r/golang: ", err)
	}
	return
}

func (rc RedditClient) RetrieveNew() (err error) {
	fmt.Print(time.Now().Format("2006-01-02 15:04:05"), " : ", "Auto retrieving new Reddit posts... ")
	var harvest reddit.Harvest = rc.Graw()
	err = rc.sendPosts(harvest.Posts)
	return
}

func (rc RedditClient) sendPosts(posts []*reddit.Post) (err error) {

	// sort "posts" base on scores
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Ups > posts[j].Ups
	})

	var i int
	var post *reddit.Post
	var leastScore int
	leastScore, err = strconv.Atoi(os.Getenv("AutoRedditLeaseScore"))
	if err != nil {
		return
	}
	for i, post = range posts {
		if i < leastScore { // filter out qualified posts
			break
		}
	}
	if i < 1 {
		fmt.Println("No new post from Reddit!")
		return
	} else {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), " : ", "Auto retrieving Hacker news posts... ")
		fmt.Printf("found %d Reddit stories.\n", i)
	}

	var mbarr []MessageBlock
	for _, post = range posts[:i] {
		var exist bool = false
		for _, existID := range rc.retrivedStoriesIds {
			if post.ID == existID {
				exist = true
				break
			}
		}
		if exist {
			continue
		} else { // send post
			rc.retrivedStoriesIds = append(rc.retrivedStoriesIds, post.ID) // add new post id to existing ones
			if strings.Contains(post.URL, "v.redd.it") {                   // video
				mbarr = rc.createVideoMsgBlock(post)
			} else if strings.Contains(post.URL, "i.redd.it") { // image
				mbarr = rc.createImageMsgBlock(post)
			} else { // post/link
				mbarr = rc.createTextMsgBlock(post)
			}

			var mbs = MessageBlocks{Blocks: mbarr}
			if flag.Lookup("test.v") == nil && Hostname != "MacBook-Pro.local" {
				err = sc.SendBlocks(mbs, rc.WebHookUrlReddit) // send the new and not published stories to slack #hacker-news
			} else {
				err = sc.SendBlocks(mbs, sc.WebHookUrlTest)
			}
			if err != nil {
				return
			}
		}
	}

	file, _ := json.Marshal(rc.retrivedStoriesIds)
	utils.WriteFile(file, redditFilename)
	return
}

func (rc RedditClient) createVideoMsgBlock(post *reddit.Post) (mbs []MessageBlock) {

	var videoLink string

	if Hostname == "MacBook-Pro.local" {
		videoLink = post.Media.RedditVideo.FallbackURL
	} else {
		var fp string = "/tmp/" + post.ID + ".mp4"
		if _, err := os.Stat(fp); os.IsNotExist(err) {
			utils.DownloadFile(post.Media.RedditVideo.FallbackURL, fp, true)
		}
		videoLink = os.Getenv("ServerIPAddr") + "/file/" + post.ID + ".mp4"
	}

	mbs = []MessageBlock{
		{Type: "divider"},
		{ // video
			Type: "section",
			Text: &ElementText{
				Type: "mrkdwn",
				Text: fmt.Sprintf("*<%s|%s>*\nups:*%d*, sub: *r/%s*\n[video]", videoLink, post.Title, post.Ups, post.Subreddit),
			},
			Accessory: &ImageAccessory{
				Type:     "image",
				ImageUrl: post.Thumbnail,
				AltText:  post.Title,
			},
		},
	}
	return
}

func (rc RedditClient) createImageMsgBlock(post *reddit.Post) (mbs []MessageBlock) {
	mbs = []MessageBlock{
		{Type: "divider"},
		{
			Type: "section",
			Text: &ElementText{
				Text: fmt.Sprintf("%s\nups: *%d*, sub: *r/%s* <image>", post.Title, post.Ups, post.Subreddit),
				Type: "mrkdwn",
			},
		},
		{ // image
			Type:     "image",
			ImageUrl: post.URL,
			AltText:  post.Title,
		},
	}
	return
}

func (rc RedditClient) createTextMsgBlock(post *reddit.Post) (mbs []MessageBlock) {
	var text string
	if strings.Contains(post.URL, post.Permalink) { // meaning this is a reddit post not news link
		text = fmt.Sprintf("*[%s] <https://reddit.com%s|%s>*\nups:*%d*, sub: *r/%s* <post>", post.ID, post.Permalink, post.Title, post.Ups, post.Subreddit)
	} else {
		text = fmt.Sprintf("*<%s|%s>*\n[<https://reddit.com%s|Reddit>] ups:*%d*, sub: *r/%s*<link>", post.URL, post.Title, post.Permalink, post.Ups, post.Subreddit)
	}
	mbs = []MessageBlock{
		{Type: "divider"},
		{ // text
			Type: "section",
			Text: &ElementText{
				Type: "mrkdwn",
				Text: text,
			},
		},
	}
	return
}
