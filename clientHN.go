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

const hnFilename string = "ids-hn.json"

func (hn HNClient) RetrieveNew(leastScoreStr string) (err error) {

	var leastScore int
	leastScore, err = strconv.Atoi(leastScoreStr)

	fmt.Print(time.Now().Format("2006-01-02 15:04:05"), " : ", "Auto retrieving new Hacker news posts... ")

	var retrivedStoriesIds []int
	_ = json.Unmarshal(utils.ReadFile(hnFilename), &retrivedStoriesIds)

	var storiesIdsList []int = hn.getStoriesIds(os.Getenv("AutoHNPostType")) // get 500 newest ids

	for {
		var edit bool = false
		for i, newId := range storiesIdsList {
			for _, existId := range retrivedStoriesIds {
				if newId == existId || i+1 == len(storiesIdsList) {
					storiesIdsList = append(storiesIdsList[:i], storiesIdsList[i+1:]...)
					edit = true
					break
				}
			}
			if edit {
				break
			}
		}
		if !edit {
			break
		}
	}

	// turn storiesIdsList into batches because it's too long
	var storiesLen int = len(storiesIdsList)
	var storiesIdsListBatches [][]int
	for i := 0; i < storiesLen/100; i++ { // turn storiesIdsList into batches
		storiesIdsListBatches = append(storiesIdsListBatches, storiesIdsList[i*100:(i+1)*100])
	}
	storiesIdsListBatches = append(storiesIdsListBatches, storiesIdsList[storiesLen-storiesLen%100:])

	// get all the story items
	var storiesItemsList []HNItem
	for _, idsBatch := range storiesIdsListBatches {
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
		fmt.Println("No qualified new post found.")
		return
	} else {
		fmt.Printf("found %d stories. ", i)
	}
	storiesIdsList = []int{}
	storiesItemsList = storiesItemsList[:i]
	for i, item = range storiesItemsList {
		storiesIdsList = append(storiesIdsList, item.Id)
	}

	// save json
	retrivedStoriesIds = append(retrivedStoriesIds, storiesIdsList...)
	j, _ := json.Marshal(retrivedStoriesIds)
	utils.WriteFile(j, hnFilename)

	var mbs MessageBlocks
	for i := 0; i < len(storiesIdsList); i++ {
		mbs, err = hn.hnStoriesToBlocks("", storiesItemsList[i:i+1], true)
		if err != nil {
			return
		}
		err = sc.SendBlocks(mbs, sc.WebHookUrlHN) // send the new and not published stories to slack #hacker-news
	}
	fmt.Println("Sent.")
	return
}

func (hn HNClient) GetHNStories(storyTypeInfo string) (msgBlocks MessageBlocks, err error) {

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
	var storiesIdsList []int = hn.getStoriesIds(storyType)
	storiesItemsList = hn.getStoriesItems(storiesIdsList)
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
		messageBlocks = append(messageBlocks, sc.CreateTextBlock(fmt.Sprintf("*%s*", storyTypeInfo), "mrkdwn"))
	}
	for _, story = range stories {
		var text string = fmt.Sprintf(
			"*<%s|%s>*\n[<%s|hn>] Score: %d, Comments: %d  @%s [%s]",
			story.Url, story.Title, fmt.Sprintf(hn.PageUrlTmplt, story.Id), story.Score, len(story.Kids), hn.parseHostname(story.Url), utils.ConvertUnixTime(story.Time),
		)
		if useDivider {
			messageBlocks = append(messageBlocks, MessageBlock{Type: "divider"})
		}
		messageBlocks = append(messageBlocks, sc.CreateTextBlock(text, "mrkdwn"))
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

func (hn HNClient) getStoriesItems(storiesIdsList []int) (storiesItemsList []HNItem) {
	var m sync.Map
	storiesItemsList = []HNItem{}

	// start := time.Now()
	defer func() { // turn m sync.Map into storiesItemsList after the process is done
		// log.Println("Execution Time: ", time.Since(start))
		for _, id := range storiesIdsList {
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
	for _, id := range storiesIdsList {
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

func (hn HNClient) getStoriesIds(storyType string) (storiesIdsList []int) {
	// top [500], new [500], best [200]
	var url string = fmt.Sprintf(hn.StoriesUrlTmplt, storyType)
	var body []byte = utils.RetrieveBytes(url)

	if err := json.Unmarshal(body, &storiesIdsList); err != nil {
		log.Fatalln(err)
	}
	return
}
