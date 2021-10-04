package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/dghubble/oauth1"
)

type TwitterClient struct{}

const convoEndpoint string = "https://api.twitter.com/2/tweets/search/recent?query=conversation_id:%s from:%s to:%s&max_results=100&expansions=author_id,in_reply_to_user_id,referenced_tweets.id&tweet.fields=in_reply_to_user_id,author_id,created_at,conversation_id"
const usersLookupEndpoint string = "https://api.twitter.com/2/users%s&user.fields=created_at,description,entities,id,location,name,pinned_tweet_id,profile_image_url,protected,url,username,verified,withheld&expansions=pinned_tweet_id&tweet.fields=attachments,author_id,conversation_id,created_at,entities,geo,id,in_reply_to_user_id,lang,possibly_sensitive,referenced_tweets,source,text,withheld"
const tweetLoopUpEndpoint string = "https://api.twitter.com/1.1/statuses/show.json?id=%s&tweet_mode=extended"
const listEndpoint string = "https://api.twitter.com/1.1/lists/statuses.json?list_id=%s&count=1000&tweet_mode=extended"
const twitterFilename string = "ids/ids-twitter.json"

// const tweetsEndpoint string = "https://api.twitter.com/2/tweets?ids=%s&tweet.fields=public_metrics,attachments,conversation_id,author_id,created_at,entities,geo,id,in_reply_to_user_id,lang,possibly_sensitive,referenced_tweets,source,text"

var TweetLists = map[string]string{
	"Makers":        "1229215345526722560",
	"Entrepreneurs": "1229216130662723584",
	"Greats":        "1310225357019074562",
	"Investors":     "1237393320378118149",
	"Physicists":    "1394817230630572034",
	"YouTubers":     "1229243949950201856",
	"Writters":      "1286864227475447808",
}

func (tc TwitterClient) UnmarshalTweet() (tweetList TweetList) {
	var bytes []byte = utils.ReadFile("data-samples/tweet.json")
	_ = json.Unmarshal(bytes, &tweetList)
	return
}

func (tc TwitterClient) LookUpTweets(ids []string) (tweets []Tweet, err error) {
	tweets = make([]Tweet, len(ids))
	var i int
	var id string
	wg := sync.WaitGroup{}
	for i, id = range ids {
		wg.Add(1)
		go func(i int, id string) {
			defer wg.Done()
			var tweet Tweet
			if tweet, err = tc.lookUpTweet(id); err != nil {
				return
			}
			tweets[i] = tweet
		}(i, id)
	}
	wg.Wait()
	return
}

func (tc TwitterClient) lookUpTweet(id string) (tweet Tweet, err error) {
	var url string = fmt.Sprintf(tweetLoopUpEndpoint, id)
	var respJson []byte
	if respJson, err = tc.SendHttpRequest(url, "v1"); err != nil {
		return
	}
	_ = json.Unmarshal(respJson, &tweet)
	utils.WriteFile(respJson, "data-samples/tweet.json")
	return
}

func (tc TwitterClient) LookUpTwitterUsers(ids []string, idType string) (respJson []byte, err error) {
	var idsString string = ids[0]
	for _, id := range ids[1:] {
		idsString = idsString + "," + id
	}
	if idType == "id" {
		if respJson, err = tc.SendHttpRequest(fmt.Sprintf(usersLookupEndpoint, fmt.Sprintf("?ids=%s", idsString)), "v2"); err != nil {
			log.Panic(err)
		}
	} else if idType == "username" {
		if respJson, err = tc.SendHttpRequest(fmt.Sprintf(usersLookupEndpoint, fmt.Sprintf("/by?usernames=%s", idsString)), "v2"); err != nil {
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
		var leastOriginalLikes int
		var mbList [][]MessageBlock
		leastOriginalLikes, _ = strconv.Atoi(os.Getenv("AutoTwitterLeastOriginalLikes"))
		mbList, err = tc.retrieveTweets(listName, leastOriginalLikes, true)
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
		if flag.Lookup("test.v") != nil { // if in test mode, only go through 1 loop
			break
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
	var listTweets, qualifiedListTweets []Tweet
	if listTweets, err = tc.GetListContent(listName); err != nil {
		return
	}

	var savedTweetIds []string
	json.Unmarshal(utils.ReadFile(twitterFilename), &savedTweetIds)

	// check if tweets are qualified
	var listTweet Tweet
	for _, listTweet = range listTweets {
		var retweet *Tweet = listTweet.Retweeted_Status
		var quoted *Tweet = listTweet.Quoted_Status
		var exist bool = false
		var leastRetweetLikes int
		leastRetweetLikes, _ = strconv.Atoi(os.Getenv("AutoTwitterLeastRetweetLikes"))
		if listTweet.Favorite_Count >= leastLikes || (retweet != nil && retweet.Favorite_Count >= leastRetweetLikes) || (quoted != nil && quoted.Favorite_Count >= leastRetweetLikes) {
			for _, savedId := range savedTweetIds {
				if listTweet.Id_Str == savedId {
					exist = true
					break
				}
			}
			if !exist {
				savedTweetIds = append(savedTweetIds, listTweet.Id_Str)
				qualifiedListTweets = append(qualifiedListTweets, listTweet)
			}
		}
	}

	var mbarr []MessageBlock
	for _, listTweet = range qualifiedListTweets {
		if mbarr, err = tc.FormatTweet(listTweet); err != nil {
			return
		}
		mbList = append(mbList, mbarr)
	}
	// save json
	if saveIDs {
		j, _ := json.Marshal(savedTweetIds)
		utils.WriteFile(j, twitterFilename)
	}

	return
}

func (tc TwitterClient) TestFormatTweet(tweet Tweet) (err error) {
	var mbarr []MessageBlock
	// var mbList [][]MessageBlock
	if mbarr, err = tc.FormatTweet(tweet); err != nil {
		return
	}
	var mbs MessageBlocks
	mbs.Blocks = append(mbs.Blocks, mbarr...)
	if err = sc.SendBlocks(mbs, os.Getenv("SlackWebHookUrlTwitter")); err != nil {
		return
	}
	return
}

func (tc TwitterClient) FormatTweet(tweet Tweet) (mbarr []MessageBlock, err error) {
	mbarr = append(mbarr, MessageBlock{Type: "divider"})
	var txt string
	var retweet *Tweet
	if tweet.Retweeted_Status == nil {
		retweet = tweet.Quoted_Status
	} else {
		retweet = tweet.Retweeted_Status
	}

	var reg *regexp.Regexp = regexp.MustCompile(`https:\/\/t.co\/([A-Za-z0-9])\w+`) // remove links like "https://t.co/se6Ys5aJ4x"
	tweet.Full_Text = reg.ReplaceAllString(tweet.Full_Text, "")
	mbarr = append(mbarr, tc.Addthumbnail(tweet.User.Profile_image_url_https, tweet.User.Screen_Name))
	if retweet != nil { // if it's a retweet
		retweet.Full_Text = reg.ReplaceAllString(retweet.Full_Text, "")
		txt = " RT"
		if tweet.Full_Text[:4] != "RT @" {
			txt = tweet.Full_Text + txt
		}
		mbarr = append(mbarr, sc.CreateTextBlock(txt, "mrkdwn", ""))
		mbarr = append(mbarr, tc.loopMediaList(tweet.Extended_Entities.Media)...)
		mbarr = append(mbarr, tc.Addthumbnail(retweet.User.Profile_image_url_https, retweet.User.Screen_Name))
		mbarr = append(mbarr, sc.CreateTextBlock(retweet.Full_Text, "mrkdwn", ""))
		txt = fmt.Sprintf(`[<https://twitter.com/%s/status/%s|tweet>] retweets: *%d*, likes: *%d*`, tweet.User.Screen_Name, tweet.Id_Str, retweet.Retweet_Count, retweet.Favorite_Count)
		mbarr = append(mbarr, sc.CreateTextBlock(txt, "mrkdwn", ""))
		mbarr = append(mbarr, tc.loopMediaList(retweet.Extended_Entities.Media)...)
	} else {
		mbarr = append(mbarr, sc.CreateTextBlock(tweet.Full_Text, "mrkdwn", ""))
		txt = fmt.Sprintf(`[<https://twitter.com/%s/status/%s|tweet>] retweets: *%d*, likes: *%d*`, tweet.User.Screen_Name, tweet.Id_Str, tweet.Retweet_Count, tweet.Favorite_Count)
		mbarr = append(mbarr, sc.CreateTextBlock(txt, "mrkdwn", ""))
		mbarr = append(mbarr, tc.loopMediaList(tweet.Extended_Entities.Media)...)
	}
	return
}

func (tc TwitterClient) Addthumbnail(thumbnailUrl, username string) MessageBlock {
	return MessageBlock{
		Type: "context",
		Elements: []*Element{{
			Type:      "image",
			Image_Url: thumbnailUrl,
			AltText:   "profile",
		}, {
			Type: "mrkdwn",
			Text: fmt.Sprintf(`<https://twitter.com/%s|@%s>`, username, username),
		}},
	}
}

func (tc TwitterClient) loopMediaList(mediaList []TweetMedia) (mbarr []MessageBlock) {
	var media TweetMedia
	for _, media = range mediaList {
		var mb MessageBlock = sc.CreateImageBlock(media.Media_Url_Https, "ok")
		mbarr = append(mbarr, mb)
		mbarr = append(mbarr, MessageBlock{
			Type: "context",
			Elements: []*Element{{
				Type:  "plain_text",
				Text:  media.Type,
				Emoji: true,
			}},
		})
	}
	return
}

func (tc TwitterClient) GetListContent(listName string) (tweets []Tweet, err error) {
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

func (tc TwitterClient) GetThread(tweetID, userID string) (err error) {
	var tweets []Tweet
	if tweets, err = tc.getThreadTweets(tweetID, userID); err != nil {
		return
	}
	if tc.sendThread(tweets); err != nil {
		return
	}
	return
}

func (tc TwitterClient) sendThread(threadList []Tweet) (err error) {
	var mbs MessageBlocks
	for _, tweet := range threadList {
		var mbarr []MessageBlock
		if mbarr, err = tc.FormatTweet(tweet); err != nil {
			return
		}
		mbs.Blocks = append(mbs.Blocks, mbarr...)
	}
	if err = sc.SendBlocks(mbs, os.Getenv("SlackWebHookUrlTest")); err != nil {
		return
	}
	return
}

func (tc TwitterClient) getThreadTweets(convoID, userID string) (tweets []Tweet, err error) {
	var url string = fmt.Sprintf(convoEndpoint, convoID, userID, userID)
	url = strings.ReplaceAll(url, " ", "%20")

	var respJson []byte
	var thread Thread
	if respJson, err = tc.SendHttpRequest(url, "v2"); err != nil {
		return
	}
	if err = json.Unmarshal(respJson, &thread); err != nil {
		return
	}
	if tweets, err = tc.sortThreadTweets(thread); err != nil {
		return
	}
	return
}

func (tc TwitterClient) sortThreadTweets(thread Thread) (tweets []Tweet, err error) {
	var sortedThreadTweets []ThreadTweetInfo
	var threadTweet, tweet ThreadTweetInfo
	var threadTweets []ThreadTweetInfo
	for _, tweet = range append(thread.Data, thread.Includes.Tweets...) {
		var exist bool = false
		for _, threadTweet = range threadTweets {
			if tweet.Id == threadTweet.Id {
				exist = true
				break
			}
		}
		if !exist {
			threadTweets = append(threadTweets, tweet)
			if tweet.Referenced_tweets == nil {
				sortedThreadTweets = append(sortedThreadTweets, tweet)
			}
		}
	}
	for i := 0; i < len(threadTweets); i++ {
		var id string = sortedThreadTweets[i].Id
		for _, threadTweet = range threadTweets {
			var referencedTweet ThreadReferencedTweet
			for _, t := range threadTweet.Referenced_tweets {
				if t.Type == "replied_to" {
					referencedTweet = t
					break
				}
			}
			if referencedTweet.Id == id {
				sortedThreadTweets = append(sortedThreadTweets, threadTweet)
				break
			}
		}
	}
	var ids []string
	for _, tt := range sortedThreadTweets {
		ids = append(ids, tt.Id)
	}

	if tweets, err = tc.LookUpTweets(ids); err != nil {
		return
	}
	return
}
