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

//error and recovery functions
func recovery() {
	if r := recover(); r != nil {
		fmt.Println(r)
		os.Exit(1)
	}
}

func main() {
	//number of args
	argsNum := uint8(len(os.Args))

	//if no args given
	switch argsNum {
	case 1:
		fmt.Println("EpubGoTerm: \n\tGo utility to print EPUB chapters into the terminal.")
		fmt.Println("Syntax:\n\tepubgot epubfile chapternumber\t# print a chapter in the EPUB\n\tepubgot epubfile\t\t# list chapters of the EPUB and print image count")
		os.Exit(0)
	}

	//recovery function
	defer recovery()

	//make arg1
	var arg1 = os.Args[1]
	//check if epub and open
	if filepath.Ext(arg1) != ".epub" {
		panic(fmt.Errorf("(1) file is not an epub: %v", arg1))
	}
	archv, err := zip.OpenReader(arg1)
	defer archv.Close()
	if err != nil {
		panic(fmt.Errorf("(1) failed to open epub: %v", arg1))
	}

	switch argsNum {
	case 2:
		//make counters
		var chpCount uint
		var imgCount uint

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
	case 3:
		//make temporary directory
		tmp, err := ioutil.TempDir(os.TempDir(), "epubgot-book*")
		defer os.RemoveAll(tmp)
		if err != nil {
			panic(fmt.Errorf("(1) could not create temporary dir: %v", tmp))
		}

		//make a list and arg2
		list := make([]string, 1, 1500)
		num, _ := strconv.Atoi(os.Args[2])
		x := uint16(num)

		//range through files and find htmls
		for _, fileInArchive := range archv.File {
			switch filepath.Ext(fileInArchive.Name) {
			case ".html", ".xhtml":
				//join path, append it to list, and get list len
				fpath := filepath.Join(tmp, fileInArchive.Name)
				list = append(list, fpath)
				var listLen = uint16(len(list) - 1)

				if listLen == x {
					//check for zipslip (https://snyk.io/research/zip-slip-vulnerability)
					if !strings.HasPrefix(fpath, filepath.Clean(tmp)+string(os.PathSeparator)) {
						panic(fmt.Errorf("(2) detected illegal file path: %v", fpath))
					}

					//stat dir of file, make if not exists
					_, err := os.Stat(filepath.Dir(fpath))
					if os.IsNotExist(err) {
						os.MkdirAll(filepath.Dir(fpath), 0777)
					}

					//open the file, make new file, and copy new to opened
					fileOpened, err := fileInArchive.Open()
					if err != nil {
						panic(fmt.Errorf("(2) failed to open html file of selected chapter: %v", fileInArchive.Name))
					}

					fileCreated, err := os.Create(fpath)
					if err != nil {
						panic(fmt.Errorf("(2) failed to create html file at temporary dir: %v", fpath))
					}

					_, err = io.Copy(fileCreated, fileOpened)
					if err != nil {
						panic(fmt.Errorf("(2) failed to copy html file to temporary dir: %v to %v", fileInArchive.Name, fpath))
					}

					//close
					fileOpened.Close()
					fileCreated.Close()

					//read selected file
					htmlFile, err := ioutil.ReadFile(list[x])
					if err != nil {
						panic(fmt.Errorf("(3) could not read html file at temporary dir: %v", list[x]))
					}

					// convert to string
					html := string(htmlFile)

					txt := convertHTMLtoText(html)

					//print result
					fmt.Println("Chapter:", x)
					fmt.Println(txt)
				}
			}
		}
		//if is 0 or more than list
		if 0 == x || x > uint16(len(list)-1) {
			panic(fmt.Errorf("(3) number is out of epub's chapter range: %v", x))
		}
	}
}

func convertHTMLtoText(fileContent string) string {
	// create expressions
	regexpTags := regexp.MustCompile(`<[^<>]+>`)
	regexpI := regexp.MustCompile(`&.+;`)
	regexpComments := regexp.MustCompile(`/*.*/`)

	newContent := fileContent

	// use expressions on string
	newContent = regexpTags.ReplaceAllString(newContent, "")
	newContent = regexpI.ReplaceAllString(newContent, " ")
	newContent = regexpComments.ReplaceAllString(newContent, "")

	return newContent
}
