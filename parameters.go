package main

type Parameters struct {
	AutoRedditLeaseScore          int    `json:"AutoRedditLeaseScore,omitempty"`
	AutoHNRenewLeastScore         int    `json:"AutoHNRenewLeastScore,omitempty"`
	AutoTwitterLeastOriginalLikes int    `json:"AutoTwitterLeastOriginalLikes,omitempty"`
	AutoTwitterLeastRetweetLikes  int    `json:"AutoTwitterLeastRetweetLikes,omitempty"`
	HNClassicDaysFromDate         int    `json:"HNClassicDaysFromDate,omitempty"`
	LatestHNClassicDate           string `json:"LatestHNClassicDate,omitempty"`
	Timezone                      string `json:"Timezone,omitempty"`
	TimeFormat                    string `json:"TimeFormat,omitempty"`
}

var (
	sc     SlackClient
	tc     TwitterClient
	dc     DiscordClient
	xk     XKCDClient
	db     Database
	Params Parameters
	hn     = HNClient{
		ItemUrlTmplt:    "https://hacker-news.firebaseio.com/v0/item/%s.json",   // "https://hacker-news.firebaseio.com/v0/item/8863.json?print=pretty"
		StoriesUrlTmplt: "https://hacker-news.firebaseio.com/v0/%sstories.json", // for finding top/new/best stories
		PageUrlTmplt:    "https://news.ycombinator.com/item?id=%d",              // link to the HN page of this story
	}
	utils      Utils
	rou        Routines
	Hostname   string
	TweetLists = map[string]string{
		"Makers":        "1229215345526722560",
		"Entrepreneurs": "1229216130662723584",
		"Greats":        "1310225357019074562",
		"Investors":     "1237393320378118149",
		"Physicists":    "1394817230630572034",
		"YouTubers":     "1229243949950201856",
		"Writters":      "1286864227475447808",
	}
	IsLocal    bool
	IsTestMode bool = false
)

const (
	paramsFilename  string = "data/parameters.json"
	hnFilename      string = "data/ids-hn.json"
	redditFilename  string = "data/ids-reddit.json"
	xkcdFilename    string = "data/ids-xkcd.json"
	twitterFilename string = "data/ids-twitter.json"

	algoliaSearchEndpoint     string = "https://hn.algolia.com/api/v1/search?query=github"                                             // http://hn.algolia.com/api/v1/search?query=...
	algoliaTimeFilterEndpoint string = "http://hn.algolia.com/api/v1/search?tags=story&numericFilters=created_at_i>%d,created_at_i<%d" // http://hn.algolia.com/api/v1/search_by_date?tags=story&numericFilters=created_at_i>X,created_at_i<Y
	RedditOAuthUrl            string = "https://oauth.reddit.com/"
	RedditTokenRetrivingUrl   string = "https://www.reddit.com/api/v1/access_token"
	convoEndpoint             string = "https://api.twitter.com/2/tweets/search/recent?query=conversation_id:%s from:%s to:%s&max_results=100&expansions=author_id,in_reply_to_user_id,referenced_tweets.id&tweet.fields=in_reply_to_user_id,author_id,created_at,conversation_id"
	usersLookupEndpoint       string = "https://api.twitter.com/2/users%s&user.fields=created_at,description,entities,id,location,name,pinned_tweet_id,profile_image_url,protected,url,username,verified,withheld&expansions=pinned_tweet_id&tweet.fields=attachments,author_id,conversation_id,created_at,entities,geo,id,in_reply_to_user_id,lang,possibly_sensitive,referenced_tweets,source,text,withheld"
	tweetLoopUpEndpoint       string = "https://api.twitter.com/1.1/statuses/show.json?id=%s&tweet_mode=extended"
	listEndpoint              string = "https://api.twitter.com/1.1/lists/statuses.json?list_id=%s&count=1000&tweet_mode=extended"
	// const tweetsEndpoint string = "https://api.twitter.com/2/tweets?ids=%s&tweet.fields=public_metrics,attachments,conversation_id,author_id,created_at,entities,geo,id,in_reply_to_user_id,lang,possibly_sensitive,referenced_tweets,source,text"
)
