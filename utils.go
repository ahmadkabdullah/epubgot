package main

import (
	"fmt"
	"os"
	"regexp"
)

func convertHTMLtoText(fileContent string) string {
	// create expressions
	regexpTags := regexp.MustCompile(`<[^<>]+>`)
	regexpI := regexp.MustCompile(`&.+;`)
	regexpComments := regexp.MustCompile(`/*.*/`)
	regexpBlankLines := regexp.MustCompile(`\n*\n`)

	newContent := fileContent

	// use expressions on string
	newContent = regexpTags.ReplaceAllString(newContent, "")
	newContent = regexpI.ReplaceAllString(newContent, " ")
	newContent = regexpBlankLines.ReplaceAllString(newContent, "\n")
	newContent = regexpComments.ReplaceAllString(newContent, "")

	return newContent
}

func printErrorAndExit() {
	if r := recover(); r != nil {
		fmt.Println(r)
		os.Exit(1)
	}
}
