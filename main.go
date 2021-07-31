package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func printErrorAndExit() {
	if r := recover(); r != nil {
		fmt.Println(r)
		os.Exit(1)
	}
}

func main() {
	//var for number of arguments
	argsNum := len((os.Args)) - 1

	//if no args given
	switch argsNum {
	case 0:
		fmt.Println("EpubGoTerm: \n\tGo utility to print EPUB chapters into the terminal.")
		fmt.Println("Use:\n\tepubgot epubfile chapternumber\t# print a chapter in the EPUB\n\tepubgot epubfile\t\t# list chapters of the EPUB and print image count")
		os.Exit(0)
	}

	//recovery function
	defer printErrorAndExit()

	//check if epub
	if filepath.Ext(os.Args[1]) != ".epub" {
		panic(fmt.Errorf("(1) file is not an epub: %v", os.Args[1]))
	}

	//open epub
	archv, err := zip.OpenReader(os.Args[1])
	defer archv.Close()
	if err != nil {
		panic(fmt.Errorf("(1) failed to open epub: %v", os.Args[1]))
	}

	switch argsNum {
	//LIST when given one argument
	case 1:
		//make counters
		var chpCount, imgCount uint

		//case file extension and act
		for _, fileInArchv := range archv.File {
			switch filepath.Ext(fileInArchv.Name) {
			case ".html", ".xhtml":
				chpCount++
				fmt.Println(chpCount, fileInArchv.Name)
			case ".jpeg", "jpg", ".png":
				imgCount++
			}
		}

		//print counters
		fmt.Println("Chapters:", chpCount, "Images:", imgCount)

	//OPEN when given two arguments
	case 2:
		//make temporary directory
		tmp, err := ioutil.TempDir(os.TempDir(), "epubgot-book*")
		defer os.RemoveAll(tmp)
		if err != nil {
			panic(fmt.Errorf("(1) could not create temporary dir: %v", tmp))
		}

		//make counter and arg2
		var htmlFileCounter int
		var foundFile bool
		wantedChapter, _ := strconv.Atoi(os.Args[2])

		//range through files and find htmls
		for _, fileInArchive := range archv.File {
			//if not html then skip
			fileExt := filepath.Ext(fileInArchive.Name)
			if fileExt != ".html" && fileExt != ".xhtml" {
				continue
			}

			//add to the counter
			htmlFileCounter++

			//if not at wanted item then skip
			if htmlFileCounter != wantedChapter {
				continue
			} else {
				foundFile = true
			}

			//set path for the target file
			targetFilePath := filepath.Join(tmp, fileInArchive.Name)

			//check path for zipslip (https://snyk.io/research/zip-slip-vulnerability)
			if !strings.HasPrefix(targetFilePath, filepath.Clean(tmp)+string(os.PathSeparator)) {
				panic(fmt.Errorf("(2) detected illegal file path: %v", targetFilePath))
			}

			//stat dir of file, make if not exists
			_, err := os.Stat(filepath.Dir(targetFilePath))
			if os.IsNotExist(err) {
				os.MkdirAll(filepath.Dir(targetFilePath), 0777)
			}

			//open the file
			fileOpened, err := fileInArchive.Open()
			defer fileOpened.Close()
			if err != nil {
				panic(fmt.Errorf("(2) failed to open html file of selected chapter: %v", fileInArchive.Name))
			}

			//make new file
			fileCreated, err := os.Create(targetFilePath)
			defer fileCreated.Close()
			if err != nil {
				panic(fmt.Errorf("(2) failed to create html file at temporary dir: %v", targetFilePath))
			}

			//copy new file to opened file
			_, err = io.Copy(fileCreated, fileOpened)
			if err != nil {
				panic(fmt.Errorf("(2) failed to copy html file to temporary dir: %v to %v", fileInArchive.Name, targetFilePath))
			}

			//read selected file
			htmlFile, err := ioutil.ReadFile(targetFilePath)
			if err != nil {
				panic(fmt.Errorf("(3) could not read html file at temporary dir: %v", targetFilePath))
			}

			// convert to string
			htmlFileAsString := string(htmlFile)

			convertedText := convertHTMLtoText(htmlFileAsString)

			//print result
			fmt.Println("Chapter:", wantedChapter)
			fmt.Println(convertedText)
		}

		//imperfect range detection
		if foundFile == false {
			panic(fmt.Errorf("(2) chapter number given is likely out of range: %v", wantedChapter))
		}
	}
}

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
