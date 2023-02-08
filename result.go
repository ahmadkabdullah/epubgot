package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func listChapters(archv *zip.ReadCloser) {
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
}

func printChapter(archv *zip.ReadCloser, wantedChapter int) {
	//make temporary directory
	tmp, err := os.MkdirTemp(os.TempDir(), "epubgot-book*")
	if err != nil {
		panic(fmt.Errorf("(1) could not create temporary dir: %v", tmp))
	}
	defer os.RemoveAll(tmp)

	//make counter and arg2
	var htmlFileCounter int
	var foundFile bool

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
		if err != nil {
			panic(fmt.Errorf("(2) failed to open html file of selected chapter: %v", fileInArchive.Name))
		}
		defer fileOpened.Close()

		//make new file
		fileCreated, err := os.Create(targetFilePath)
		if err != nil {
			panic(fmt.Errorf("(2) failed to create html file at temporary dir: %v", targetFilePath))
		}
		defer fileCreated.Close()

		//copy new file to opened file
		_, err = io.Copy(fileCreated, fileOpened)
		if err != nil {
			panic(fmt.Errorf("(2) failed to copy html file to temporary dir: %v to %v", fileInArchive.Name, targetFilePath))
		}

		//read selected file
		htmlFile, err := os.ReadFile(targetFilePath)
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
	if !foundFile {
		panic(fmt.Errorf("(2) chapter number given is likely out of range: %v", wantedChapter))
	}
}
