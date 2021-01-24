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
func errorExists(err error) bool {
	if err == nil {
		return false
	}
	return true
}

func main() {
	//number of args
	argsNum := uint8(len(os.Args))

	//if no args given
	switch argsNum {
	case 1:
		fmt.Println("EpubGoTerminal: \n\tGo utility to print EPUB chapters into the terminal.")
		fmt.Println("Syntax:\n\tepubgot epubfile chapternumber\t# print a chapter in the EPUB\n\tepubgot epubfile\t\t# list chapters of the EPUB and print image count")
		fmt.Println("Version: 1.3")
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
	if errorExists(err) {
		panic(fmt.Errorf("(1) failed to open epub: %v", arg1))
	}

	switch argsNum {
	case 2:
		//make counters
		var chpCount uint16
		var imgCount uint16
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
		if errorExists(err) {
			panic(fmt.Errorf("(1) could not create temporary dir: %v", tmp))
		}

		//make a list and arg2
		list := make([]string, 1, 1500)
		num, _ := strconv.Atoi(os.Args[2])
		x := uint16(num)

		//range through files and find htmls
		for _, fileInArchv := range archv.File {
			switch filepath.Ext(fileInArchv.Name) {
			case ".html", ".xhtml":
				//join path, append it to list, and get list len
				fpath := filepath.Join(tmp, fileInArchv.Name)
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
					fileOpened, err := fileInArchv.Open()
					if errorExists(err) {
						panic(fmt.Errorf("(2) failed to open html file of selected chapter: %v", fileInArchv.Name))
					}
					fileCreated, err := os.Create(fpath)
					if errorExists(err) {
						panic(fmt.Errorf("(2) failed to create html file at temporary dir: %v", fpath))
					}
					_, err = io.Copy(fileCreated, fileOpened)
					if errorExists(err) {
						panic(fmt.Errorf("(2) failed to copy html file to temporary dir: %v to %v", fileInArchv.Name, fpath))
					}
					//close
					fileOpened.Close()
					fileCreated.Close()

					//read selected file
					htmlFile, err := ioutil.ReadFile(list[x])
					if errorExists(err) {
						panic(fmt.Errorf("(3) could not read html file at temporary dir: %v", list[x]))
					}
					html := string(htmlFile)
					//convert selected file
					var reg = regexp.MustCompile(`<[^<>]+>`)
					txt := reg.ReplaceAllString(html, "")
					var ireg = regexp.MustCompile(`&.+;`)
					txt = ireg.ReplaceAllString(txt, " ")
					var creg = regexp.MustCompile(`/*.*/`)
					txt = creg.ReplaceAllString(txt, "")

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
