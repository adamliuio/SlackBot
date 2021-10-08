package main

import (
	"encoding/json"
	"fmt"
	"log"
	urlUtils "net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

type HNClient struct {
	ItemUrlTmplt    string
	StoriesUrlTmplt string
	PageUrlTmplt    string
}

type HNItem struct {
	Id          int    // The item's unique id.
	Deleted     bool   // true if the item is deleted.
	Type        string // The type of item. One of "job", "story", "comment", "poll", or "pollopt".
	By          string // The username of the item's author.
	Descendants int    // In the case of stories or polls, the total comment count.
	Kids        []int  // The ids of the item's comments, in ranked display order.
	Parent      int    // The comment's parent: either another comment or the relevant story.
	Score       int    // The story's score, or the votes for a pollopt.
	Time        int64  // Creation date of the item, in Unix Time.
	Title       string // The title of the story, poll or job. HTML.
	Url         string // The URL of the story.
	Text        string // The comment, story or poll text. HTML.
	Dead        bool   // true if the item is dead.
	Poll        int    // The pollopt's associated poll.
	Parts       []int  // A list of related pollopts, in display order.
}

type HNAlgoliaSearchResults struct {
	Hits []HNAlgoliaSearchResult `json:"hits,omitempty"`
}

type HNAlgoliaSearchResult struct {
	Created_at   string `json:"created_at,omitempty"`
	Created_at_i int64  `json:"created_at_i,omitempty"`
	Title        string `json:"title,omitempty"`
	Url          string `json:"url,omitempty"`
	Points       int    `json:"points,omitempty"`
	Num_comments int    `json:"num_comments,omitempty"`
	ObjectID     string `json:"objectID,omitempty"` // story item id
}

func (hn HNClient) AutoHNClassic() {
	var results HNAlgoliaSearchResults
	var err error
	if results, err = hn.RetrieveHNClassic(); err != nil {
		if err.Error() == "invalid character 'S' looking for beginning of value" {
			hn.AutoHNClassic()
			return
		} else {
			utils.DealWithError(err)
		}
	}
	if err = hn.classicsFormatData(results); err != nil {
		utils.DealWithError(err)
	}
}

func (hn HNClient) RetrieveHNClassic() (results HNAlgoliaSearchResults, err error) {
	var dayBegin, dayEnd time.Time
	var layoutISO string = "2006-01-02"
	if dayBegin, err = time.Parse(layoutISO, Params.LatestHNClassicDate); err != nil {
		return
	}
	for {
		dayEnd = dayBegin.AddDate(0, 0, Params.HNClassicDaysFromDate)
		var url string = fmt.Sprintf(algoliaTimeFilterEndpoint, dayBegin.Unix(), dayEnd.Unix())
		var respBody []byte
		if respBody, err = utils.HttpRequest("GET", nil, url, nil); err != nil {
			return
		}
		if err = json.Unmarshal(respBody, &results); err != nil {
			return
		}

		sort.Slice(results.Hits, func(i, j int) bool {
			return results.Hits[i].Points > results.Hits[j].Points
		})
		var hasQualified bool = false
		for _, item := range results.Hits {
			if item.Points > Params.AutoHNRenewLeastScore {
				hasQualified = true
				break
			}
		}
		if hasQualified {
			Params.LatestHNClassicDate = dayEnd.Format(layoutISO)
			j, _ := json.Marshal(Params)
			utils.WriteFile(j, paramsFilename)
			return
		} else {
			dayBegin = dayEnd
		}
		if Hostname == "MacBook-Pro.local" {
			utils.WriteFile(respBody, "data-samples/t.json")
		}
	}
}

func (hn HNClient) classicsFormatData(results HNAlgoliaSearchResults) (err error) {
	var story HNAlgoliaSearchResult
	for _, story = range results.Hits {
		var mbarr = []MessageBlock{}
		mbarr = append(mbarr, MessageBlock{Type: "divider"})
		var text string = fmt.Sprintf(
			"*<%s|%s>*\n[<%s|hn>] Score: %d, Comments: %d\n@%s [%s]",
			story.Url, story.Title, fmt.Sprintf("https://news.ycombinator.com/item?id=%s", story.ObjectID), story.Points,
			story.Num_comments, hn.parseHostname(story.Url), utils.ConvertUnixTime(story.Created_at_i),
		)
		mbarr = append(mbarr, sc.CreateTextBlock(text, "mrkdwn", ""))
		if err = sc.SendBlocks(MessageBlocks{Blocks: mbarr}, os.Getenv("SlackWebHookUrlHNClassics")); err != nil { // send the new and not published stories to slack #hacker-news
			return
		}
	}
	return
}

func (hn HNClient) AutoRetrieveNew() (err error) {
	var str string = "found "
	for _, s := range []string{"top", "new", "best"} {
		var i int
		if i, err = hn._retrieveNew(s); err != nil {
			return
		}
		str = str + fmt.Sprintf("|%d %s| ", i, s)
	}
	fmt.Println(str + "HN stories.")
	return
}

func (hn HNClient) _retrieveNew(autoHNPostType string) (i int, err error) {

	var leastScore int = Params.AutoHNRenewLeastScore

	var savedStoriesIds []int
	_ = json.Unmarshal(utils.ReadFile(hnFilename), &savedStoriesIds)

	var newIdsList []int
	var _idsList []int
	if _idsList, err = hn.getStoriesIds(autoHNPostType); err != nil { // get 500 newest ids
		return
	}

	for _, newId := range _idsList {
		var exist bool = false
		for _, existId := range savedStoriesIds {
			if newId == existId {
				exist = true
				break
			}
		}
		if !exist {
			newIdsList = append(newIdsList, newId)
		}
	}

	// turn newIdsList into batches because it's too long
	var storiesLen int = len(newIdsList)
	var newIdsListBatches [][]int
	for i := 0; i < storiesLen/100; i++ { // turn newIdsList into batches
		newIdsListBatches = append(newIdsListBatches, newIdsList[i*100:(i+1)*100])
	}
	newIdsListBatches = append(newIdsListBatches, newIdsList[storiesLen-storiesLen%100:])

	// get all the story items
	var storiesItemsList []HNItem
	for _, idsBatch := range newIdsListBatches {
		storiesItemsList = append(storiesItemsList, hn.getStoriesItems(idsBatch)...)
	}

	// sort "storiesItemsList" base on scores
	sort.Slice(storiesItemsList, func(i, j int) bool {
		return storiesItemsList[i].Score > storiesItemsList[j].Score
	})

	// eliminate stories that scored less than 350
	var item HNItem
	for i, item = range storiesItemsList {
		if item.Score < leastScore {
			break
		}
	}
	storiesItemsList = storiesItemsList[:i]

	// save json
	for i, item = range storiesItemsList {
		savedStoriesIds = append(savedStoriesIds, item.Id)
	}
	var mbs MessageBlocks
	for i = 0; i < len(storiesItemsList); i++ {
		if mbs, err = hn.formatData("", storiesItemsList[i:i+1], true); err != nil {
			return
		}
		if err = sc.SendBlocks(mbs, os.Getenv("SlackWebHookUrlHN")); err != nil { // send the new and not published stories to slack #hacker-news
			return
		}
	}
	j, _ := json.Marshal(savedStoriesIds)
	utils.WriteFile(j, hnFilename)
	return
}

func (hn HNClient) RetrieveByCommand(storyTypeInfo string) (mbs MessageBlocks, err error) {
	var storyType string
	var storiesRange []int

	if storyType, storiesRange, err = regexStoryTypeRange(storyTypeInfo); err != nil { // parsing storyType & storiesRange
		mbs = MessageBlocks{Text: err.Error()}
		return
	}

	var stories []HNItem
	if stories, err = hn.getStories(storyType, storiesRange); err != nil {
		mbs = MessageBlocks{Text: err.Error()}
		return
	}
	mbs, err = hn.formatData(storyTypeInfo, stories, false)
	return
}

func (hn HNClient) getStories(storyType string, storiesRange []int) (storiesItemsList []HNItem, err error) {
	// top [500], new [500], best [200]
	if !strings.Contains("top/new/best", storyType) {
		err = fmt.Errorf(`the <story type> "%s" you put in is invalid, should be one if <top/new/best>`, storyType)
		return
	}
	var newIdsList []int
	if newIdsList, err = hn.getStoriesIds(storyType); err != nil {
		return
	}
	storiesItemsList = hn.getStoriesItems(newIdsList)
	sort.Slice(storiesItemsList, func(i, j int) bool {
		return storiesItemsList[i].Score > storiesItemsList[j].Score
	})
	storiesItemsList = storiesItemsList[storiesRange[0]:storiesRange[1]]
	return
}

func (hn HNClient) formatData(storyTypeInfo string, stories []HNItem, useDivider bool) (mbs MessageBlocks, err error) {
	var story HNItem
	var mbarr []MessageBlock
	if storyTypeInfo != "" {
		mbarr = append(mbarr, sc.CreateTextBlock(fmt.Sprintf("*%s*", storyTypeInfo), "mrkdwn", ""))
	}
	for _, story = range stories {
		var text string = fmt.Sprintf(
			"*<%s|%s>*\n[<%s|hn>] Score: %d, Comments: %d\n@%s [%s]",
			story.Url, story.Title, fmt.Sprintf(hn.PageUrlTmplt, story.Id), story.Score,
			len(story.Kids), hn.parseHostname(story.Url), utils.ConvertUnixTime(story.Time),
		)
		if useDivider {
			mbarr = append(mbarr, MessageBlock{Type: "divider"})
		}
		mbarr = append(mbarr, sc.CreateTextBlock(text, "mrkdwn", ""))
	}

	mbs = MessageBlocks{Blocks: mbarr}
	return
}

func (hn HNClient) parseHostname(hostname string) string {
	var err error
	var u *urlUtils.URL
	if u, err = urlUtils.Parse(hostname); err != nil {
		return fmt.Sprintln("url has issue:", err.Error())
	}
	return strings.ReplaceAll(u.Hostname(), "www.", "")
}

func (hn HNClient) getStoriesItems(newIdsList []int) (storiesItemsList []HNItem) {
	var m sync.Map
	storiesItemsList = []HNItem{}

	// start := time.Now()
	defer func() { // turn m sync.Map into storiesItemsList after the process is done
		// log.Println("Execution Time: ", time.Since(start))
		for _, id := range newIdsList {
			var item HNItem
			var itemIntf interface{}
			var ok bool
			itemIntf, ok = m.Load(id)
			if !ok {
				log.Fatalf("id: %d is no ok, detail: %+v\n", id, item)
			}
			var b []byte
			var err error
			if b, err = json.Marshal(itemIntf); err != nil {
				log.Fatalln(err)
			}
			if err = json.Unmarshal(b, &item); err != nil {
				log.Fatalln(err)
			}

			storiesItemsList = append(storiesItemsList, item)
		}
	}()

	// starting concurrent processes that retrieve hn news items simultaneously
	wg := sync.WaitGroup{}
	for _, id := range newIdsList {
		wg.Add(1)
		go func(id int) {
			var hn HNItem = hn.getItemById(hn.ItemUrlTmplt, id)
			m.Store(id, hn)
			wg.Done()
		}(id)
	}
	wg.Wait()
	return
}

func (hn HNClient) getStoriesIds(storyType string) (newIdsList []int, err error) {
	// top [500], new [500], best [200]
	var url string = fmt.Sprintf(hn.StoriesUrlTmplt, storyType)
	var body []byte
	if body, err = utils.HttpRequest("GET", nil, url, nil); err != nil {
		log.Fatalln(err)
	}

	if err = json.Unmarshal(body, &newIdsList); err != nil {
		return
	}
	return
}

func (hn HNClient) getItemById(formatStr string, id int) (item HNItem) {
	var url string = fmt.Sprintf(formatStr, id)
	var body []byte
	var err error
	if body, err = utils.HttpRequest("GET", nil, url, nil); err != nil {
		log.Panic(err)
	}
	if err = json.Unmarshal(body, &item); err != nil {
		log.Panic(err)
	}
	return
}
