package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
)

func regexStoryTypeRange(storyTypeInfo string) (storyType string, storyRange []int, err error) {

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
			log.Fatalln(err)
		}
	} else if len(numMatches) == 2 {
		var num string = storyTypeInfo[numMatches[0][0]:numMatches[0][1]]
		storyRange[0], err = strconv.Atoi(num)
		if err != nil {
			log.Fatalln(err)
		}
		num = storyTypeInfo[numMatches[1][0]:numMatches[1][1]]
		storyRange[1], err = strconv.Atoi(num)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		err = fmt.Errorf(`the command ("%s") seems to have more or less than 2 numbers, the format should either be something like "/hn top 10" or "/hn top 10-20"`, storyTypeInfo)
		return
	}
	return
}
