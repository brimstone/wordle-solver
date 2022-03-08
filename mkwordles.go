// +build generate
//go:generate go run mkwordles.go

package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func getJSPath() string {
	resp, err := http.Get("https://www.nytimes.com/games/wordle/index.html")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	i := strings.Index(string(body), "main.")
	if i < 0 {
		return ""
	}
	url := body[i:]
	j := strings.Index(string(url), "\"")
	return string(url[:j])
}

func getWordsFromJS(url string) []string {
	resp, err := http.Get(url)
	if err != nil {
		return []string{}
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	i := strings.Index(string(body), "Ma=")
	if i < 0 {
		return []string{}
	}
	body = body[i+4:]

	i = strings.Index(string(body), "]")
	body = body[:i]

	return strings.Split(strings.ReplaceAll(string(body), "\"", ""), ",")
}

func main() {
	js := getJSPath()
	if js == "" {
		fmt.Println("JS path not found")
		os.Exit(1)
	}

	js = "https://www.nytimes.com/games/wordle/" + js

	words := getWordsFromJS(js)
	if len(words) == 0 {
		fmt.Println("Unable to parse word list from website")
		os.Exit(1)
	}

	err := ioutil.WriteFile("wordles.txt", []byte(strings.Join(words, "\n")), 0777)
	if err != nil {
		fmt.Println("Error writing wordles")
		os.Exit(1)
	}
}
