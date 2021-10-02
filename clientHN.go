package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
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
	Time        int    // Creation date of the item, in Unix Time.
	Title       string // The title of the story, poll or job. HTML.
	Url         string // The URL of the story.
	Text        string // The comment, story or poll text. HTML.
	Dead        bool   // true if the item is dead.
	Poll        int    // The pollopt's associated poll.
	Parts       []int  // A list of related pollopts, in display order.
}

const hnFilename string = "ids/ids-hn.json"

func (hn HNClient) AutoRetrieveNew() (err error) {
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), ":", "Auto retrieving Hacker news posts... ")
	for _, s := range []string{"top", "new", "best"} {
		err = hn._retrieveNew(s)
		if err != nil {
			return
		}
	}
	return
}

func (hn HNClient) _retrieveNew(autoHNPostType string) (err error) {

	var leastScore int
	leastScore, err = strconv.Atoi(os.Getenv("AutoHNLeaseScore"))

	var savedStoriesIds []int
	_ = json.Unmarshal(utils.ReadFile(hnFilename), &savedStoriesIds)

	var newIdsList []int
	var _idsList []int = hn.getStoriesIds(autoHNPostType) // get 500 newest ids

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
	var i int
	var item HNItem
	for i, item = range storiesItemsList {
		if item.Score < leastScore {
			break
		}
	}
	if i < 1 {
		fmt.Printf("No new HN %s post w/ score > %s found.\n", autoHNPostType, os.Getenv("AutoHNLeaseScore"))
		return
	} else {
		fmt.Printf("found %d %s HN stories.\n", i, autoHNPostType)
	}
	storiesItemsList = storiesItemsList[:i]

	// save json
	for i, item = range storiesItemsList {
		savedStoriesIds = append(savedStoriesIds, item.Id)
	}
	var mbs MessageBlocks
	for i := 0; i < len(storiesItemsList); i++ {
		mbs, err = hn.hnStoriesToBlocks("", storiesItemsList[i:i+1], true)
		if err != nil {
			return
		}
		err = sc.SendBlocks(mbs, os.Getenv("SlackWebHookUrlHN")) // send the new and not published stories to slack #hacker-news
		if err != nil {
			return
		}
	}
	fmt.Println("Sent.")
	j, _ := json.Marshal(savedStoriesIds)
	utils.WriteFile(j, hnFilename)
	return
}

func (hn HNClient) RetrieveByCommand(storyTypeInfo string) (msgBlocks MessageBlocks, err error) {

	var storyType string
	var storiesRange []int

	storyType, storiesRange, err = regexStoryTypeRange(storyTypeInfo) // parsing storyType & storiesRange
	if err != nil {
		msgBlocks = MessageBlocks{Text: err.Error()}
		return
	}

	var stories []HNItem
	stories, err = hn.getStories(storyType, storiesRange)
	if err != nil {
		msgBlocks = MessageBlocks{Text: err.Error()}
		return
	}
	msgBlocks, err = hn.hnStoriesToBlocks(storyTypeInfo, stories, false)
	return
}

func (hn HNClient) getStories(storyType string, storiesRange []int) (storiesItemsList []HNItem, err error) {
	// top [500], new [500], best [200]
	if !strings.Contains("top/new/best", storyType) {
		err = fmt.Errorf(`the <story type> "%s" you put in is invalid, should be one if <top/new/best>`, storyType)
		return
	}
	var newIdsList []int = hn.getStoriesIds(storyType)
	storiesItemsList = hn.getStoriesItems(newIdsList)
	sort.Slice(storiesItemsList, func(i, j int) bool {
		return storiesItemsList[i].Score > storiesItemsList[j].Score
	})
	storiesItemsList = storiesItemsList[storiesRange[0]:storiesRange[1]]
	return
}

func (hn HNClient) hnStoriesToBlocks(storyTypeInfo string, stories []HNItem, useDivider bool) (msgBlocks MessageBlocks, err error) {
	var story HNItem
	var messageBlocks []MessageBlock
	if storyTypeInfo != "" {
		messageBlocks = append(messageBlocks, sc.CreateTextBlock(fmt.Sprintf("*%s*", storyTypeInfo), "mrkdwn", ""))
	}
	for _, story = range stories {
		var text string = fmt.Sprintf(
			"*<%s|%s>*\n[<%s|hn>] Score: %d, Comments: %d\n@%s [%s]",
			story.Url, story.Title, fmt.Sprintf(hn.PageUrlTmplt, story.Id), story.Score, len(story.Kids), hn.parseHostname(story.Url), utils.ConvertUnixTime(story.Time),
		)
		if useDivider {
			messageBlocks = append(messageBlocks, MessageBlock{Type: "divider"})
		}
		messageBlocks = append(messageBlocks, sc.CreateTextBlock(text, "mrkdwn", ""))
	}

	msgBlocks = MessageBlocks{Blocks: messageBlocks}
	return
}

func (hn HNClient) parseHostname(hostname string) string {
	u, err := url.Parse(hostname)
	if err != nil {
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
			b, err := json.Marshal(itemIntf)
			if err != nil {
				log.Fatalln(err)
			}
			err = json.Unmarshal(b, &item)
			if err != nil {
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
			var hn HNItem = utils.GetItemById(hn.ItemUrlTmplt, id)
			m.Store(id, hn)
			wg.Done()
		}(id)
	}
	wg.Wait()
	return
}

func (hn HNClient) getStoriesIds(storyType string) (newIdsList []int) {
	// top [500], new [500], best [200]
	var url string = fmt.Sprintf(hn.StoriesUrlTmplt, storyType)
	var body []byte = utils.RetrieveBytes(url, nil)

	if err := json.Unmarshal(body, &newIdsList); err != nil {
		log.Fatalln(err)
	}
	return
}
