package main

import (
	"crypto/rand"
	b64 "encoding/base64"
	"encoding/json"
	"flag"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"
)

func TestThread(t *testing.T) {
	tc.GetThread("1444268274267693057", "15735804")
}

func TestRegex(t *testing.T) {
	var txt string = string(utils.ReadFile("data-samples/tweetUrls.txt"))
	var reg *regexp.Regexp = regexp.MustCompile(`https:\/\/t.co\/([A-Za-z0-9])\w+`)
	var res []string = reg.FindAllString(txt, -1)
	var redirects = make(map[string]string)

	var wg sync.WaitGroup
	for _, url := range res {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			finalurl, _, err := utils.CheckUrl(url)
			if err != nil {
				log.Fatal(err)
			}
			if strings.Contains(finalurl, "twitter.com") {
				redirects[url] = finalurl
			}
		}(url)
	}
	wg.Wait()

	for url := range redirects {
		txt = strings.ReplaceAll(txt, url, "")
	}
	t.Log(txt)
}

func TestYolo(t *testing.T) { // pulls a tweet and send to slack
	var tweets []Tweet
	var err error
	if tweets, err = tc.LookUpTweets([]string{"1449922164744982529"}); err != nil {
		t.Fatal(err)
	}
	_ = tweets
	var tweet Tweet
	_ = json.Unmarshal(utils.ReadFile("data-samples/tweet.json"), &tweet)
	// t.Fatalf("%+v\n", tweet)
	if err := tc.TestFormatTweet(tweet); err != nil {
		t.Fatal(err)
	}
}

func TestFormatTweet(t *testing.T) {
	var tweet Tweet
	_ = json.Unmarshal(utils.ReadFile("data-samples/tweet.json"), &tweet)
	// t.Fatalf("%+v\n", tweet.Extended_Entities.Media[0].Media_Url_Https == tweet.Retweeted_Status.Extended_Entities.Media[0].Media_Url_Https)
	if err := tc.TestFormatTweet(tweet); err != nil {
		t.Fatal(err)
	}
}

func TestLookUpTweets(t *testing.T) {
	var tweets []Tweet
	var err error
	if tweets, err = tc.LookUpTweets([]string{"1444268274267693057"}); err != nil {
		t.Fatal(err)
	}
	t.Logf("tweet: %+v\n", tweets)
}

func TestLookUpTwitterUsers(t *testing.T) {
	// curl --location --request GET 'https://api.twitter.com/2/users/2244994946'
	var respJson []byte
	var err error
	if respJson, err = tc.LookUpTwitterUsers([]string{"bbourque"}, "username"); err != nil {
		log.Fatalln(err)
	}
	// var respJsonStr string = utils.PrettyJsonString(respJson)
	var respJsonStr string = string(respJson)
	t.Log(respJsonStr)

	var user TwitterUser
	_ = json.Unmarshal(respJson, &user)
	t.Logf("TwitterUser: %+v\n", user)
}

func TestUnmarshalTweet(t *testing.T) {
	var tweetList TweetList = tc.UnmarshalTweet()
	t.Logf("TweetListData: %+v\n", tweetList)
}

func TestEndpoints(t *testing.T) {
	// var url string = "https://api.twitter.com/2/users?ids=783214,2244994945,6253282,495309159,172020392,95731075,2548985366,277761722,17874544,300392950,87532773,372575989,3260518932,121291606,158079127,3282859598,103770785,586198217,216531294,1526228120,222953824,1603818258,2548979088,2244983430,1347713256,376825877,6844292,738118115595165697,738118487122419712,218984871,2550997820,1159458169,2296297326,234489024,3873936134,2228891959,791978718,427475002,1194267639100723200,1168976680867762177,905409822,738115375477362688,88723966,1049385226424786944,284201599,1705676064,2861317614,3873965293,1244731491088809984,4172587277,717465714357972992,862314223,2551000568,2548977510,1159274324&user.fields=created_at,description,entities,id,location,name,pinned_tweet_id,profile_image_url,protected,url,username,verified,withheld&expansions=pinned_tweet_id&tweet.fields=attachments,author_id,conversation_id,created_at,entities,geo,id,in_reply_to_user_id,lang,non_public_metrics,organic_metrics,possibly_sensitive,promoted_metrics,referenced_tweets,source,text,withheld"
	var url string = "https://api.twitter.com/1.1/lists/members.json?list_id=1229215345526722560"
	var respJson []byte
	var err error
	if respJson, err = tc.SendHttpRequest(url, "v2"); err != nil {
		log.Fatalln(err)
	}
	var respJsonStr string = utils.PrettyJsonString(respJson)
	// respJson = string(body)
	t.Log(respJsonStr)
}

func TestOAuthV1(t *testing.T) {
	var url string = "https://api.twitter.com/1.1/lists/statuses.json?list_id=1229215345526722560"
	var respJson []byte
	var err error
	if respJson, err = tc.SendHttpRequest(url, "v1"); err != nil {
		log.Fatalln(err)
	}
	t.Log(string(respJson))
}

func Test32Byte(t *testing.T) {
	token := make([]byte, 32)
	rand.Read(token)
	t.Log(token)
	sEnc := b64.StdEncoding.EncodeToString(token)
	t.Log(sEnc)
}

func TestGetListContent(t *testing.T) {
	var tweets []Tweet
	var err error
	if tweets, err = tc.GetListContent("Makers"); err != nil {
		return
	}
	t.Log(len(tweets))
}

func TestListContentMarshal(t *testing.T) {
	var tweets []Tweet
	_ = json.Unmarshal(utils.ReadFile("data-samples/list-statuses.json"), &tweets)

	t.Log(len(tweets))
	for _, tweet := range tweets {
		t.Logf("%+v\n", tweet)
		if tweet.Retweeted_Status != nil {
			t.Logf("%+v\n", tweet.Retweeted_Status)
		}
		// t.Log(tweet.Text)
		// t.Log(fmt.Sprintf("https://twitter.com/Trekhleb/status/%s", tweet.Id))
		// t.Log(tweet.User.Screen_Name)
		// t.Log(tweet.User.Id)
	}
}

func TestRetrieveByCommand(t *testing.T) {
	var msgBlocks MessageBlocks
	var err error
	msgBlocks, err = tc.RetrieveByCommand("Makers 5000")
	if err != nil {
		t.Fatal(err)
	}
	err = sc.SendBlocks(msgBlocks, os.Getenv("SlackWebHookUrlTest"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestTest(t *testing.T) {
	t.Logf("%+v\n", flag.Lookup("test.v") == nil) // not in test mode
	t.Logf("%+v\n", flag.Arg(0))
}

func TestAutoRetrieveNew(t *testing.T) {
	var err error = tc.AutoRetrieveNew()
	if err != nil {
		panic(err)
	}
}
