package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
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
	respJson = tc.SendHttpRequest(fmt.Sprintf(tweetEndpoint, idsString))
	return
}

func (tc TwitterClient) LookUpTwitterUsers(ids []string, idType string) (respJson []byte) {
	var idsString string = ids[0]
	for _, id := range ids[1:] {
		idsString = idsString + "," + id
	}
	if idType == "id" {
		respJson = tc.SendHttpRequest(fmt.Sprintf(usersByIdEndpoint, idsString))
	} else if idType == "username" {
		respJson = tc.SendHttpRequest(fmt.Sprintf(usersByUsernameEndpoint, idsString))
	} else {
		log.Fatalf("id type: %s is wrong.", idType)
	}
	return
}

func (tc TwitterClient) SendHttpRequest(url string) (body []byte) {
	var headers = [][]string{{"Authorization", fmt.Sprintf("Bearer %s", os.Getenv(("TwitterBearerToken")))}}
	body = utils.RetrieveBytes(url, headers)
	return
}
