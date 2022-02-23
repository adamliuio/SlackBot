package main

import (
	"encoding/json"
	"fmt"
	"log"
	urlUtils "net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
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
			j, _ := json.MarshalIndent(Params, "", "    ")
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
		fmt.Printf("Retrieving %s HN stories.\n", s)
		var i int
		if i, err = hn.retrieveNew(s); err != nil {
			log.Panicln(err)
		}
		str = str + fmt.Sprintf("|%d %s| ", i, s)
	}
	fmt.Println(str + "HN stories.")
	sc.SendPlainText(fmt.Sprintf("Sent %s HN stories.", str), os.Getenv("SlackWebHookUrlTest"))
	return
}

func (hn HNClient) retrieveNew(autoHNPostType string) (i int, err error) {

	var leastScore int = Params.AutoHNRenewLeastScore

	var newIdsList []string
	if newIdsList, err = hn.getStoriesIds(autoHNPostType); err != nil { // get 500 newest ids
		return
	}

	// turn newIdsList into batches because it's too long multi-threading
	var storiesLen int = len(newIdsList)
	var newIdsListBatches [][]string
	for i := 0; i < storiesLen/100; i++ { // turn newIdsList into batches
		newIdsListBatches = append(newIdsListBatches, newIdsList[i*100:(i+1)*100])
	}
	newIdsListBatches = append(newIdsListBatches, newIdsList[storiesLen-storiesLen%100:])

	var storiesItemsList []HNItem
	var qualifiedSavedItems []SavedItem

	for _, idsBatch := range newIdsListBatches {
		var batchItemsList []HNItem = hn.getStoriesItems(idsBatch) // get items of this batch base on batch ids

		var item HNItem
		for _, item = range batchItemsList {
			if item.Score >= leastScore { // only deal with qualified items
				var newId string = fmt.Sprint(item.Id)
				var returnedItem SavedItem = db.QueryRow(newId) // check if exists
				if returnedItem.Platform == "HackerNews" {      // if exists
					continue
				} else {
					storiesItemsList = append(storiesItemsList, item)
					qualifiedSavedItems = append(qualifiedSavedItems, SavedItem{Id: newId, Platform: "HackerNews"})
				}
			}
		}
	}
	if len(qualifiedSavedItems) > 0 {
		db.InsertRows(qualifiedSavedItems)
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
	return
}

func (hn HNClient) RetrieveByCommand(storyTypeInfo string) (mbs MessageBlocks, err error) {
	var storyType string
	var storiesRange []int

	if storyType, storiesRange, err = hn.regexStoryTypeRange(storyTypeInfo); err != nil { // parsing storyType & storiesRange
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

func (hn HNClient) regexStoryTypeRange(storyTypeInfo string) (storyType string, storyRange []int, err error) {

	matchAll := regexp.MustCompile(`[\D\W]+\s\d+(-\d+)?`).MatchString(storyTypeInfo)  // match the format of the entire string
	wordMatch := regexp.MustCompile(`[^\d\W]+`).FindAllStringIndex(storyTypeInfo, -1) // match words
	numMatches := regexp.MustCompile(`\d+`).FindAllStringIndex(storyTypeInfo, -1)     // match numbers

	if !matchAll {
		err = fmt.Errorf(`command ("%s") wrong, should either be something like "/hn top 10" or "/hn top 10-20"`, storyTypeInfo)
		return
	}

	storyType = storyTypeInfo[wordMatch[0][0]:wordMatch[0][1]]
	storyRange = []int{0, 10}

	if len(numMatches) == 1 { // if there's values in the string, which is separated by " "
		var num string = storyTypeInfo[numMatches[0][0]:numMatches[0][1]]
		storyRange[1], err = strconv.Atoi(num)
		if err != nil {
			utils.DealWithError(err)
		}
	} else if len(numMatches) == 2 {
		var num string = storyTypeInfo[numMatches[0][0]:numMatches[0][1]]
		storyRange[0], err = strconv.Atoi(num)
		if err != nil {
			utils.DealWithError(err)
		}
		num = storyTypeInfo[numMatches[1][0]:numMatches[1][1]]
		storyRange[1], err = strconv.Atoi(num)
		if err != nil {
			utils.DealWithError(err)
		}
	} else {
		err = fmt.Errorf(`the command ("%s") seems to have more or less than 2 numbers, the format should either be something like "/hn top 10" or "/hn top 10-20"`, storyTypeInfo)
		return
	}
	return
}

func (hn HNClient) getStories(storyType string, storiesRange []int) (storiesItemsList []HNItem, err error) {
	// top [500], new [500], best [200]
	if !strings.Contains("top/new/best", storyType) {
		err = fmt.Errorf(`the <story type> "%s" you put in is invalid, should be one if <top/new/best>`, storyType)
		return
	}
	var newIdsList []string
	if newIdsList, err = hn.getStoriesIds(storyType); err != nil {
		return
	}
	storiesItemsList = hn.getStoriesItems(newIdsList) // retrieve the actual data of each item
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

func (hn HNClient) getStoriesItems(newIdsList []string) (storiesItemsList []HNItem) {
	var m sync.Map
	storiesItemsList = []HNItem{}

	// start := time.Now()
	defer func() { // turn m sync.Map into storiesItemsList after the process is done
		// log.Println("Execution Time: ", time.Since(start))
		var id string
		for _, id = range newIdsList {
			var item HNItem
			var itemIntf interface{}
			var ok bool
			itemIntf, ok = m.Load(id)
			if !ok {
				utils.DealWithError(fmt.Errorf("id: %s is no ok, detail: %+v", id, item))
			}
			var b []byte
			var err error
			if b, err = json.Marshal(itemIntf); err != nil {
				utils.DealWithError(err)
			}
			if err = json.Unmarshal(b, &item); err != nil {
				utils.DealWithError(err)
			}

			storiesItemsList = append(storiesItemsList, item)
		}
	}()

	// starting concurrent processes that retrieve hn news items simultaneously
	wg := sync.WaitGroup{}
	var id string
	for _, id = range newIdsList {
		wg.Add(1)
		go func(id string) {
			var hn HNItem = hn.getItemById(hn.ItemUrlTmplt, id)
			m.Store(id, hn)
			wg.Done()
		}(id)
	}
	wg.Wait()
	return
}

func (hn HNClient) getStoriesIds(storyType string) (newIdsList []string, err error) {
	// top [500], new [500], best [200]
	var url string = fmt.Sprintf(hn.StoriesUrlTmplt, storyType)
	var body []byte
	if body, err = utils.HttpRequest("GET", nil, url, nil); err != nil {
		utils.DealWithError(err)
	}

	var lst []int
	if err = json.Unmarshal(body, &lst); err != nil {
		return
	}
	for _, id := range lst {
		newIdsList = append(newIdsList, fmt.Sprint(id))
	}
	return
}

func (hn HNClient) getItemById(formatStr string, id string) (item HNItem) {
	var url string = fmt.Sprintf(formatStr, id)
	var body []byte
	var err error
	if body, err = utils.HttpRequest("GET", nil, url, nil); err != nil {
		utils.DealWithError(err)
	}
	if err = json.Unmarshal(body, &item); err != nil {
		utils.DealWithError(err)
	}
	return
}
