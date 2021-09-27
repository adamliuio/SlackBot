package main

type Tweet struct {
	Data   TweetData      `json:"data,omitempty"`
	Errors []TwitterError `json:"errors,omitempty"`
	Title  string         `json:"title,omitempty"`  // for error
	Detail string         `json:"detail,omitempty"` // for error
	Type   string         `json:"type,omitempty"`   // for error
}

type TweetData struct {
	Entities           TweetEntities    `json:"entities,omitempty"`
	Author_Id          string           `json:"author_id,omitempty"`
	Text               string           `json:"text,omitempty"`
	Possibly_Sensitive bool             `json:"possibly_sensitive,omitempty"`
	Id                 string           `json:"id,omitempty"`
	Source             string           `json:"source,omitempty"`
	Lang               string           `json:"lang,omitempty"`
	Created_At         string           `json:"created_at,omitempty"`
	TweetAttachments   TweetAttachments `json:"attachments,omitempty"`
}

type TweetAttachments struct {
	Media_Keys []string `json:"media_keys,omitempty"`
}
type TweetEntities struct {
	Urls             []TweetUrl        `json:"urls,omitempty"`
	TweetAnnotations []TweetAnnotation `json:"annotations,omitempty"`
	Description      []Hashtag         `json:"description,omitempty"`
	Url              TwitterUserUrls   `json:"url,omitempty"`
}

type TwitterUserUrls struct {
	Urls []TweetUrl `json:"urls,omitempty"`
}

// "entities": {  // user
// 	"url": {
// 		"urls": [
// 			{
// 				"start": 0,
// 				"end": 23,
// 				"url": "https://t.co/DAtOo6uuHk",
// 				"expanded_url": "https://about.twitter.com/",
// 				"display_url": "about.twitter.com"
// 			}
// 		]
// 	}

type Hashtag struct {
	Start int    `json:"start,omitempty"`
	End   int    `json:"end,omitempty"`
	Tag   string `json:"tag,omitempty"`
}

type TweetUrl struct {
	Start        int          `json:"start,omitempty"`
	End          int          `json:"end,omitempty"`
	Url          string       `json:"url,omitempty"`
	Expanded_Url string       `json:"expanded_url,omitempty"`
	Display_Url  string       `json:"display_url,omitempty"`
	Images       []TweetImage `json:"images,omitempty"`
	Status       int          `json:"status,omitempty"`
	Title        string       `json:"title,omitempty"`
	Description  string       `json:"description,omitempty"`
	Unwound_Url  string       `json:"unwound_url,omitempty"`
}

type TweetImage struct {
	Url    string `json:"url,omitempty"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

type TweetAnnotation struct {
	Start           int    `json:"start,omitempty"`
	End             int    `json:"end,omitempty"`
	Probability     int    `json:"probability,omitempty"`
	Type            string `json:"type,omitempty"`
	Normalized_Text string `json:"normalized_text,omitempty"`
}

type TwitterError struct {
	Value         string             `json:"value,omitempty"`
	Detail        string             `json:"detail,omitempty"`
	Title         string             `json:"title,omitempty"`
	Resource_Type string             `json:"resource_type,omitempty"`
	Parameter     TwitterErrorParams `json:"parameter,omitempty"`
	Resource_Id   string             `json:"resource_id,omitempty"`
	Type          string             `json:"type,omitempty"`
	Parameters    string             `json:"parameters,omitempty"`
	Message       string             `json:"message,omitempty"`
}

type TwitterErrorParams struct {
	Id []string `json:"id,omitempty"`
}

// = = =

type TwitterUser struct {
	Data   []UserData     `json:"data,omitempty"`
	Errors []TwitterError `json:"errors,omitempty"`
	Title  string         `json:"title,omitempty"`  // for error
	Detail string         `json:"detail,omitempty"` // for error
	Type   string         `json:"type,omitempty"`   // for error
}

type UserData struct {
	Url               string           `json:"url,omitempty"`
	Name              string           `json:"name,omitempty"`
	Profile_Image_Url string           `json:"profile_image_url,omitempty"`
	Entities          TweetEntities    `json:"entities,omitempty"`
	Pinned_Tweet_Id   string           `json:"pinned_tweet_id,omitempty"`
	Verified          bool             `json:"verified,omitempty"`
	Description       string           `json:"description,omitempty"`
	Protected         bool             `json:"protected,omitempty"`
	Created_At        string           `json:"created_at,omitempty"`
	Username          string           `json:"username,omitempty"`
	Location          string           `json:"location,omitempty"`
	Id                string           `json:"id,omitempty"`
	Includes          UserDataIncludes `json:"includes,omitempty"`
}

type UserDataIncludes struct {
	Tweets []TweetData `json:"includes,omitempty"`
}
