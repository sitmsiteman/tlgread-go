package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"unicode"
)

// cleanString removes high-bit characters and non-printable bytes
func cleanString(s string) string {
	return strings.Map(func(r rune) rune {
		// Keep only printable ASCII to ensure compatibility with grep and text editors
		if unicode.IsPrint(r) && r < 128 {
			return r
		}
		return -1
	}, s)
}

func main() {

	fPath := flag.String("f", "authtab.dir", "filename")
	flag.Parse()

	if *fPath == "" {
		log.Fatal("Usage: ./readauth -f authtab.dir")
	}

	f, err := os.Open(*fPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	filename := *fPath

	// 1. Read the binary directory
	data, err := ioutil.ReadFile(filename)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	// 2. Optimized Regex for TLG-E format
	// Stops capturing at '&' or any byte >= 0x80 to prevent including binary delimiters
	re := regexp.MustCompile(`TLG(\d{4}).*?&1([^&\x80-\xff]+)(?:&([^&\x80-\xffTLG]+))?`)
	matches := re.FindAllSubmatch(data, -1)

	fmt.Printf("Parsed %d authors.\n", len(matches))
	fmt.Println("----------------------------------------")

	for _, m := range matches {
		id := string(m[1])
		author := strings.TrimSpace(cleanString(string(m[2])))
		epithet := ""
		if len(m) > 3 {
			epithet = strings.TrimSpace(cleanString(string(m[3])))
		}

		// Format the line
		var line string
		if epithet != "" {
			line = fmt.Sprintf("ID: %s | Author: %s (%s)", id, author, epithet)
		} else {
			line = fmt.Sprintf("ID: %s | Author: %s", id, author)
		}

		// Output to both terminal and file
		fmt.Println(line)
	}
}
