### About

Epubgot is a utility to output EPUB into terminal as plaintext. Written in Golang, it does two things:

- List content: display how many chapters there are to read and how many images you'll miss.
- Output chapter: unzips EPUBs, selects an HTML file, converts it to text and prints to the terminal.


### Usage

Show info and help.
```sh
epubgot
```

List all chapters and number of images in the book.
```sh
epubgot programmingingo.epub
```

Output the 12th chapter of the book
```sh
epubgot programmingingo.epub 12
```

Output all chapters of the book
```sh
epubgot programmingingo.epub all
```
