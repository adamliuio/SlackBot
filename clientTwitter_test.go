package main

import (
	"encoding/json"
	"testing"
)

var tc = TwitterClient{}

func TestLookUpTweet(t *testing.T) {
	// curl --location -H "Authorization: Bearer AAAAAAAAAAAAAAAAAAAAACo3UAEAAAAApRD0U8wmPjq94WTSNXvS7FC7EfA%3DilVpIVZpStoSpHvir6o8WX5Tgw69dHapVjcf8F8UUzLk9BlqVm" --request GET 'https://api.twitter.com/2/tweets/1275828087666679809?tweet.fields=attachments,author_id,created_at,entities,geo,id,in_reply_to_user_id,lang,possibly_sensitive,referenced_tweets,source,text,withheld'
	var respJson []byte = tc.LookUpTweet([]string{"1275828087666679809", "1440976651085565952"})
	var respJsonStr string = utils.PrettyJsonString(respJson)
	// respJson = string(body)
	t.Log(respJsonStr)

	var tweet Tweet
	_ = json.Unmarshal(respJson, &tweet)
	t.Logf("Tweet: %+v\n", tweet)
}

func TestLookUpTwitterUsers(t *testing.T) {
	// curl --location --request GET 'https://api.twitter.com/2/users/2244994946'
	var respJson []byte = tc.LookUpTwitterUsers([]string{"twitter", "twitterdev", "twitterapi", "googledevs"}, "username") // 783214,2244994945,twitter,twitterdev,twitterapi,googledevs
	var respJsonStr string = utils.PrettyJsonString(respJson)
	// respJson = string(body)
	t.Log(respJsonStr)

	var user TwitterUser
	_ = json.Unmarshal(respJson, &user)
	t.Logf("TwitterUser: %+v\n", user)
}

func TestUnmarshalTweet(t *testing.T) {
	var tweet Tweet = tc.UnmarshalTweet()
	t.Logf("Tweet: %+v\n", tweet)
}

func TestEndpoints(t *testing.T) {
	var url string = "https://api.twitter.com/2/users?ids=783214,2244994945,6253282,495309159,172020392,95731075,2548985366,277761722,17874544,300392950,87532773,372575989,3260518932,121291606,158079127,3282859598,103770785,586198217,216531294,1526228120,222953824,1603818258,2548979088,2244983430,1347713256,376825877,6844292,738118115595165697,738118487122419712,218984871,2550997820,1159458169,2296297326,234489024,3873936134,2228891959,791978718,427475002,1194267639100723200,1168976680867762177,905409822,738115375477362688,88723966,1049385226424786944,284201599,1705676064,2861317614,3873965293,1244731491088809984,4172587277,717465714357972992,862314223,2551000568,2548977510,1159274324&user.fields=created_at,description,entities,id,location,name,pinned_tweet_id,profile_image_url,protected,url,username,verified,withheld&expansions=pinned_tweet_id&tweet.fields=attachments,author_id,conversation_id,created_at,entities,geo,id,in_reply_to_user_id,lang,non_public_metrics,organic_metrics,possibly_sensitive,promoted_metrics,referenced_tweets,source,text,withheld"
	var respJson []byte = tc.SendHttpRequest(url)
	var respJsonStr string = utils.PrettyJsonString(respJson)
	// respJson = string(body)
	t.Log(respJsonStr)
}

// const userEndpoint string = "https://api.twitter.com/2/users/by/username/TwitterDec"
