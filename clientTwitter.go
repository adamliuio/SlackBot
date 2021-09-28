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

const tweetEndpoint string = "https://api.twitter.com/2/tweets?ids=%s?tweet.fields=attachments,conversation_id,author_id,created_at,entities,geo,id,in_reply_to_user_id,lang,possibly_sensitive,referenced_tweets,source,text,withheld"
const usersByUsernameEndpoint string = "https://api.twitter.com/2/users/by?usernames=%s&user.fields=created_at,description,entities,id,location,name,pinned_tweet_id,profile_image_url,protected,url,username,verified,withheld&expansions=pinned_tweet_id&tweet.fields=attachments,author_id,conversation_id,created_at,entities,geo,id,in_reply_to_user_id,lang,possibly_sensitive,referenced_tweets,source,text,withheld"
const usersByIdEndpoint string = "https://api.twitter.com/2/users?ids=%s&user.fields=created_at,description,entities,id,location,name,pinned_tweet_id,profile_image_url,protected,url,username,verified,withheld&expansions=pinned_tweet_id&tweet.fields=attachments,author_id,conversation_id,created_at,entities,geo,id,in_reply_to_user_id,lang,possibly_sensitive,referenced_tweets,source,text,withheld"
const listEndpoint string = "https://api.twitter.com/1.1/lists/statuses.json?list_id=%s&count=1000"
const twitterFilename string = "ids-twitter.json"

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

func (tc TwitterClient) LookUpTweet(ids []string) (respJson []byte) {
	var idsString string = ids[0]
	for _, id := range ids[1:] {
		idsString = idsString + "," + id
	}
	respJson = tc.SendHttpRequest(fmt.Sprintf(tweetEndpoint, idsString), "v2")
	return
}

func (tc TwitterClient) LookUpTwitterUsers(ids []string, idType string) (respJson []byte) {
	var idsString string = ids[0]
	for _, id := range ids[1:] {
		idsString = idsString + "," + id
	}
	if idType == "id" {
		respJson = tc.SendHttpRequest(fmt.Sprintf(usersByIdEndpoint, idsString), "v2")
	} else if idType == "username" {
		respJson = tc.SendHttpRequest(fmt.Sprintf(usersByUsernameEndpoint, idsString), "v2")
	} else {
		log.Fatalf("id type: %s is wrong.", idType)
	}
	return
}

func (tc TwitterClient) SendHttpRequest(url, version string) (body []byte) {
	if version == "v1" {
		body = tc.oauth1Request(url)
	} else if version == "v2" {
		var headers = [][]string{{"Authorization", fmt.Sprintf("Bearer %s", os.Getenv("TwitterBearerToken"))}}
		body = utils.RetrieveBytes(url, headers)
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
	var savedTweetsIds []int
	json.Unmarshal(utils.ReadFile(twitterFilename), &savedTweetsIds)

	var mbs MessageBlocks
	for listName, _ := range TweetLists {
		var limit int
		var qualifiedTweetIds []int
		limit, _ = strconv.Atoi(os.Getenv("AutoTwitterLeaseLikes"))
		mbs, qualifiedTweetIds, err = tc.retrieveTweets(listName, limit)
		if err != nil {
			return
		}
		savedTweetsIds = append(savedTweetsIds, qualifiedTweetIds...)
		if flag.Lookup("test.v") == nil { // if this is not in test mode
			err = sc.SendBlocks(mbs, os.Getenv("WebHookUrlTwitter"))
		} else { // if is test mode
			err = sc.SendBlocks(mbs, os.Getenv("WebHookUrlTest"))
		}
		if err != nil {
			return
		}
	}

	// save json
	j, _ := json.Marshal(savedTweetsIds)
	utils.WriteFile(j, twitterFilename)
	return
}

func (tc TwitterClient) RetrieveByCommand(cmdTxt string) (mbs MessageBlocks, err error) { // /twt listname limit
	var limit int
	var listName, numStr string
	var fields []string = strings.Fields(cmdTxt)
	listName = fields[0]
	numStr = fields[1]
	limit, _ = strconv.Atoi(numStr)
	mbs, _, err = tc.retrieveTweets(listName, limit)
	return
}

func (tc TwitterClient) retrieveTweets(listName string, limit int) (mbs MessageBlocks, qualifiedTweetIds []int, err error) {
	var tweets []ListTweet = tc.GetListContent(listName)

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
	for _, tweet := range tweets {
		if tweet.Favorite_Count >= limit {
			qualifiedTweetIds = append(qualifiedTweetIds, tweet.Id)
			mbarr, err = tc.formatTweet(tweet)
			if err != nil {
				return
			}
			mbs.Blocks = append(mbs.Blocks, mbarr...)
		} else if tweet.Retweeted_Status != nil && tweet.Retweeted_Status.Favorite_Count >= limit {
			qualifiedTweetIds = append(qualifiedTweetIds, tweet.Id)
			mbarr, err = tc.formatTweet(tweet)
			if err != nil {
				return
			}
			mbs.Blocks = append(mbs.Blocks, mbarr...)
		}
	}

	return
}

func (tc TwitterClient) formatTweet(tweet ListTweet) (mbarr []MessageBlock, err error) {
	var txt, imageUrl string
	if tweet.Retweeted_Status != nil { // if it's a retweet
		txt = fmt.Sprintf(
			`<https://twitter.com/%s|@%s> RT:
<https://twitter.com/%s|@%s>: %s
[<https://twitter.com/%s/status/%s|tweet>] retweets: *%d*, likes: *%d*`,
			tweet.User.Screen_Name, tweet.User.Screen_Name, tweet.Retweeted_Status.User.Screen_Name, tweet.Retweeted_Status.User.Screen_Name, tweet.Retweeted_Status.Text,
			tweet.User.Screen_Name, tweet.Id_Str, tweet.Retweeted_Status.Retweet_Count, tweet.Retweeted_Status.Favorite_Count,
		)
	} else {
		txt = fmt.Sprintf(
			`<https://twitter.com/%s|@%s>: %s
[<https://twitter.com/%s/status/%s|tweet>] retweets: *%d*, likes: *%d*`,
			tweet.User.Screen_Name, tweet.User.Screen_Name, tweet.Text,
			tweet.User.Screen_Name, tweet.Id_Str, tweet.Retweet_Count, tweet.Favorite_Count,
		)
	}
	var medias []TweetMedia
	if len(tweet.Extended_Entities.Media) > 0 {
		medias = tweet.Extended_Entities.Media
	} else if tweet.Retweeted_Status != nil && len(tweet.Retweeted_Status.Extended_Entities.Media) > 0 {
		medias = tweet.Retweeted_Status.Extended_Entities.Media
	}
	for _, media := range medias {
		imageUrl = media.Media_Url_Https
		if media.Type == "photo" {
			txt = txt + " [pic]"
		} else if media.Type == "video" {
			txt = txt + " [vid]"
		}
	}
	var tweetBlock MessageBlock = sc.CreateTextBlock(txt, "mrkdwn", imageUrl)
	mbarr = append(mbarr, MessageBlock{Type: "divider"})
	mbarr = append(mbarr, tweetBlock)
	return
}

func (tc TwitterClient) GetListContent(listName string) (tweets []ListTweet) {
	var url string = fmt.Sprintf(listEndpoint, TweetLists[listName])
	var respJson []byte = tc.SendHttpRequest(url, "v1")
	_ = json.Unmarshal(respJson, &tweets)
	return
}
