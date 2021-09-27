package main

import (
	"encoding/json"
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

const tweetEndpoint string = "https://api.twitter.com/2/tweets/%s?tweet.fields=attachments,author_id,created_at,entities,geo,id,in_reply_to_user_id,lang,possibly_sensitive,referenced_tweets,source,text,withheld"
const usersByUsernameEndpoint string = "https://api.twitter.com/2/users/by?usernames=%s&user.fields=created_at,description,entities,id,location,name,pinned_tweet_id,profile_image_url,protected,url,username,verified,withheld&expansions=pinned_tweet_id&tweet.fields=attachments,author_id,conversation_id,created_at,entities,geo,id,in_reply_to_user_id,lang,possibly_sensitive,referenced_tweets,source,text,withheld"
const usersByIdEndpoint string = "https://api.twitter.com/2/users?ids=%s&user.fields=created_at,description,entities,id,location,name,pinned_tweet_id,profile_image_url,protected,url,username,verified,withheld&expansions=pinned_tweet_id&tweet.fields=attachments,author_id,conversation_id,created_at,entities,geo,id,in_reply_to_user_id,lang,possibly_sensitive,referenced_tweets,source,text,withheld"

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
	// var url string = "https://api.twitter.com/1.1/lists/statuses.json?list_id=1229215345526722560"
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

func (tc TwitterClient) RetrieveNew() (err error) {
	return
}

func (tc TwitterClient) RetrieveByCommand(cmdTxt string) (msgBlocks MessageBlocks, err error) {
	var listName, numStr string
	var limit int
	var fields []string = strings.Fields(cmdTxt)
	listName = fields[0]
	numStr = fields[1]
	limit, _ = strconv.Atoi(numStr)

	// var tweets []ListTweet = tc.GetListContent(listName)
	_ = listName
	var tweets []ListTweet
	_ = json.Unmarshal(utils.ReadFile("data-samples/list-statuses.json"), &tweets)

	// sort "storiesItemsList" base on scores
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

	// log.Println(len(tweets))
	// log.Fatalln(limit)
	for _, tweet := range tweets[:limit] {
		var txt string
		if tweet.Retweeted_Status != nil {
			txt = fmt.Sprintf(
				`<https://twitter.com/%s|@%s>: "%s"
				<https://twitter.com/%s|@%s>: "%s"`,
				tweet.User.Screen_Name, tweet.User.Screen_Name, tweet.Text,
				tweet.Retweeted_Status.User.Screen_Name, tweet.Retweeted_Status.User.Screen_Name, tweet.Retweeted_Status.Text,
			)
		} else {
			txt = fmt.Sprintf(
				`<https://twitter.com/%s|@%s>:\n"%s"`,
				tweet.User.Screen_Name, tweet.User.Screen_Name, tweet.Text,
			)
		}
		msgBlocks.Blocks = append(msgBlocks.Blocks, MessageBlock{Type: "divider"})
		msgBlocks.Blocks = append(msgBlocks.Blocks, sc.CreateTextBlock(txt, "mrkdwn"))
	}
	return
}

func (tc TwitterClient) GetListContent(listName string) (tweets []ListTweet) {
	var lists = map[string]string{
		"Makers":        "1229215345526722560",
		"Entrepreneurs": "1229216130662723584",
		"Greats":        "1310225357019074562",
		"Investors":     "1237393320378118149",
		"Physicists":    "1394817230630572034",
		"YouTubers":     "1229243949950201856",
		"Writters":      "1286864227475447808",
	}
	var url string = fmt.Sprintf("https://api.twitter.com/1.1/lists/statuses.json?list_id=%s&count=1000", lists[listName])

	var respJson []byte = tc.SendHttpRequest(url, "v1")
	_ = json.Unmarshal(respJson, &tweets)
	return
}
