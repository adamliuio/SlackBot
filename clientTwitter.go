package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/dghubble/oauth1"
)

type TwitterClient struct{}

const usersByUsernameEndpoint string = "https://api.twitter.com/2/users/by?usernames=%s&user.fields=created_at,description,entities,id,location,name,pinned_tweet_id,profile_image_url,protected,url,username,verified,withheld&expansions=pinned_tweet_id&tweet.fields=attachments,author_id,conversation_id,created_at,entities,geo,id,in_reply_to_user_id,lang,possibly_sensitive,referenced_tweets,source,text,withheld"
const usersByIdEndpoint string = "https://api.twitter.com/2/users?ids=%s&user.fields=created_at,description,entities,id,location,name,pinned_tweet_id,profile_image_url,protected,url,username,verified,withheld&expansions=pinned_tweet_id&tweet.fields=attachments,author_id,conversation_id,created_at,entities,geo,id,in_reply_to_user_id,lang,possibly_sensitive,referenced_tweets,source,text,withheld"
const listEndpoint string = "https://api.twitter.com/1.1/lists/statuses.json?list_id=%s&count=1000"
const twitterFilename string = "ids/ids-twitter.json"

var TweetLists = map[string]string{
	"Makers":        "1229215345526722560",
	"Entrepreneurs": "1229216130662723584",
	"Greats":        "1310225357019074562",
	"Investors":     "1237393320378118149",
	"Physicists":    "1394817230630572034",
	"YouTubers":     "1229243949950201856",
	"Writters":      "1286864227475447808",
}

func (tc TwitterClient) UnmarshalTweet() (tweet Tweet) {
	var bytes []byte = utils.ReadFile("data-samples/tweet.json")
	_ = json.Unmarshal(bytes, &tweet)
	return
}

// func (tc TwitterClient) LookUpTweet(ids []string) (respJson []byte) {
// 	var idsString string = ids[0]
// 	for _, id := range ids[1:] {
// 		idsString = idsString + "," + id
// 	}
// 	respJson = tc.SendHttpRequest(fmt.Sprintf(tweetEndpoint, idsString), "v2")
// 	return
// }

func (tc TwitterClient) LookUpTwitterUsers(ids []string, idType string) (respJson []byte, err error) {
	var idsString string = ids[0]
	for _, id := range ids[1:] {
		idsString = idsString + "," + id
	}
	if idType == "id" {
		if respJson, err = tc.SendHttpRequest(fmt.Sprintf(usersByIdEndpoint, idsString), "v2"); err != nil {
			log.Panic(err)
		}
	} else if idType == "username" {
		if respJson, err = tc.SendHttpRequest(fmt.Sprintf(usersByUsernameEndpoint, idsString), "v2"); err != nil {
			log.Panic(err)
		}
	} else {
		log.Fatalf("id type: %s is wrong.", idType)
	}
	return
}

func (tc TwitterClient) SendHttpRequest(url, version string) (body []byte, err error) {
	if version == "v1" {
		body = tc.oauth1Request(url)
	} else if version == "v2" {
		var headers = [][]string{{"Authorization", fmt.Sprintf("Bearer %s", os.Getenv("TwitterBearerToken"))}}
		if body, err = utils.HttpRequest("GET", nil, url, headers); err != nil {
			log.Panic(err)
		}
	}
	return
}

func (tc TwitterClient) oauth1Request(url string) (body []byte) {
	config := oauth1.NewConfig(os.Getenv("TwitterApiKey"), os.Getenv("TwitterApiKeySecret"))
	token := oauth1.NewToken(os.Getenv("TwitterAccessToken"), os.Getenv("TwitterAccessTokenSecret"))
	httpClient := config.Client(oauth1.NoContext, token)
	resp, err := httpClient.Get(url)
	if err != nil {
		log.Panic(err)
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err)
	}
	return
}

func (tc TwitterClient) AutoRetrieveNew() (err error) {

	for listName := range TweetLists {
		var leastLikes int
		var mbList [][]MessageBlock
		leastLikes, _ = strconv.Atoi(os.Getenv("AutoTwitterLeaseLikes"))
		mbList, err = tc.retrieveTweets(listName, leastLikes, true)
		if err != nil {
			return
		}
		var i int
		var mb []MessageBlock
		for i, mb = range mbList {
			var mbs MessageBlocks
			if i == 0 {
				mbs.Blocks = []MessageBlock{{Type: "header", Text: &ElementText{Type: "plain_text", Text: listName}}}
			}
			mbs.Blocks = append(mbs.Blocks, mb...)
			err = sc.SendBlocks(mbs, os.Getenv("SlackWebHookUrlTwitter"))
			if err != nil {
				return
			}
		}
	}
	return
}

func (tc TwitterClient) RetrieveByCommand(cmdTxt string) (mbs MessageBlocks, err error) { // /twt listname limit
	var leastLikes int
	var listName, numStr string
	var fields []string = strings.Fields(cmdTxt)
	listName = fields[0]
	numStr = fields[1]
	leastLikes, _ = strconv.Atoi(numStr)
	var mbList [][]MessageBlock
	mbList, err = tc.retrieveTweets(listName, leastLikes, false)
	if err != nil {
		return
	}
	var i int
	var mb []MessageBlock
	for i, mb = range mbList {
		if i == 0 {
			mbs.Blocks = []MessageBlock{{Type: "header", Text: &ElementText{Type: "plain_text", Text: listName}}}
		}
		mbs.Blocks = append(mbs.Blocks, mb...)
		if err != nil {
			return
		}
		if i == 1 {
			break
		}
	}
	return
}

func (tc TwitterClient) retrieveTweets(listName string, leastLikes int, saveIDs bool) (mbList [][]MessageBlock, err error) {
	var tweets []ListTweet
	if tweets, err = tc.GetListContent(listName); err != nil {
		return
	}

	// sort "tweets" base on scores
	sort.Slice(tweets, func(i, j int) bool {
		var tweetI, tweetJ ListTweet = tweets[i], tweets[j]
		var tweetIcount, tweetJcount int
		if tweetI.Retweeted_Status != nil {
			tweetIcount = tweetI.Favorite_Count + tweetI.Retweeted_Status.Favorite_Count
		} else {
			tweetIcount = tweetI.Favorite_Count
		}
		if tweetJ.Retweeted_Status != nil {
			tweetJcount = tweetJ.Favorite_Count + tweetJ.Retweeted_Status.Favorite_Count
		} else {
			tweetJcount = tweetJ.Favorite_Count
		}
		return tweetIcount > tweetJcount
	})

	var mbarr []MessageBlock
	var savedTweetIds []int
	json.Unmarshal(utils.ReadFile(twitterFilename), &savedTweetIds)
	for _, tweet := range tweets {
		var exist bool = false
		for _, savedTweetId := range savedTweetIds {
			if tweet.Id == savedTweetId {
				exist = true
				break
			}
		}
		if exist {
			continue
		}
		if tweet.Favorite_Count >= leastLikes || tweet.Retweeted_Status != nil && tweet.Retweeted_Status.Favorite_Count >= leastLikes {
			savedTweetIds = append(savedTweetIds, tweet.Id)
			if mbarr, err = tc.formatTweet(tweet); err != nil {
				return
			}
			mbList = append(mbList, mbarr)
		}
	}
	// save json
	if saveIDs {
		j, _ := json.Marshal(savedTweetIds)
		utils.WriteFile(j, twitterFilename)
	}
	return
}

func (tc TwitterClient) formatTweet(tweet ListTweet) (mbarr []MessageBlock, err error) {
	mbarr = append(mbarr, MessageBlock{Type: "divider"})
	var txt string
	if tweet.Retweeted_Status != nil { // if it's a retweet
		txt = fmt.Sprintf(`<https://twitter.com/%s|@%s> RT:`, tweet.User.Screen_Name, tweet.User.Screen_Name)
		mbarr = append(mbarr, sc.CreateTextBlock(txt, "mrkdwn", ""))

		txt = fmt.Sprintf(`<https://twitter.com/%s|@%s>: %s`, tweet.Retweeted_Status.User.Screen_Name, tweet.Retweeted_Status.User.Screen_Name, tweet.Retweeted_Status.Text)
		mbarr = append(mbarr, sc.CreateTextBlock(txt, "mrkdwn", ""))

		txt = fmt.Sprintf(`[<https://twitter.com/%s/status/%s|tweet>] retweets: *%d*, likes: *%d*`, tweet.User.Screen_Name, tweet.Id_Str, tweet.Retweeted_Status.Retweet_Count, tweet.Retweeted_Status.Favorite_Count)
		mbarr = append(mbarr, sc.CreateTextBlock(txt, "mrkdwn", ""))
		for _, media := range tweet.Retweeted_Status.Extended_Entities.Media {
			var mb MessageBlock = sc.CreateImageBlock(media.Media_Url_Https, "")
			mbarr = append(mbarr, mb)
			if media.Type == "photo" {
				txt = txt + " [pic]"
			} else if media.Type == "video" {
				txt = txt + " [vid]"
			}
		}
	} else {
		txt = fmt.Sprintf(`<https://twitter.com/%s|@%s>: %s`, tweet.User.Screen_Name, tweet.User.Screen_Name, tweet.Text)
		mbarr = append(mbarr, sc.CreateTextBlock(txt, "mrkdwn", ""))

		txt = fmt.Sprintf(`[<https://twitter.com/%s/status/%s|tweet>] retweets: *%d*, likes: *%d*`, tweet.User.Screen_Name, tweet.Id_Str, tweet.Retweet_Count, tweet.Favorite_Count)
		mbarr = append(mbarr, sc.CreateTextBlock(txt, "mrkdwn", ""))
		for _, media := range tweet.Extended_Entities.Media {
			var mb MessageBlock = sc.CreateImageBlock(media.Media_Url_Https, "")
			mbarr = append(mbarr, mb)
			if media.Type == "photo" {
				txt = txt + " [pic]"
			} else if media.Type == "video" {
				txt = txt + " [vid]"
			}
		}
	}
	return
}

func (tc TwitterClient) GetListContent(listName string) (tweets []ListTweet, err error) {
	var url string = fmt.Sprintf(listEndpoint, TweetLists[listName])
	var respJson []byte
	if respJson, err = tc.SendHttpRequest(url, "v1"); err != nil {
		return
	}
	_ = json.Unmarshal(respJson, &tweets)
	if flag.Lookup("test.v") != nil { // if in test mode
		utils.WriteFile(respJson, "data-samples/list-statuses.json")
	}
	return
}
