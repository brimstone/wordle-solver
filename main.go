package main

import (
	"bufio"
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
)

//go:embed wordles.txt
var words embed.FS

type Guess struct {
	Word   string
	Result string
}

func (g *Guess) WordA() []string {
	return strings.Split(g.Word, "")
}
func (g *Guess) ResultA() []string {
	if g.Result == "" {
		g.Result = "_____"
	}
	return strings.Split(g.Result, "")
}

func getGrays(guesses []Guess) (ret []string) {
	h := make(map[string]bool)
	for _, w := range guesses {
		for i, l := range w.WordA() {
			if w.ResultA()[i] == "_" {
				h[l] = true
			}
		}
	}
	for k := range h {
		ret = append(ret, k)
	}
	sort.Strings(ret)

	return
}

func getYellows(guesses []Guess) []string {
	var yellows []string
	var greens []string
	for _, g := range guesses {
		for i := 0; i < 5; i++ {
			if g.ResultA()[i] == "y" {
				yellows = append(yellows, g.WordA()[i])
				continue
			}
			if g.ResultA()[i] == "g" {
				greens = append(greens, g.WordA()[i])
				continue
			}
		}
	}
	for _, green := range greens {
		for i := len(yellows) - 1; i >= 0; i-- {
			if yellows[i] == green {
				yellows = append(yellows[:i], yellows[i:]...)
			}
		}
	}
	return yellows
}

func buildRegex(guesses []Guess) (ret string) {
	ret = "^"
	grays := strings.Join(getGrays(guesses), "")
	for i := 0; i < 5; i++ {
		p := ""
		y := ""
		for _, g := range guesses {
			if g.ResultA()[i] == "g" {
				p = g.WordA()[i]
				ret += p
				break
			}
			// TODO handle when a letter is both a yellow and a green
			if g.ResultA()[i] == "y" {
				y = g.WordA()[i]
			}
		}
		if p != "" {
			continue
		}
		ret += "[^" + y + grays + "]"
	}
	ret += "$"
	return
}

func getBestGuess(candidates []string) string {
	score := make([]int, len(candidates))
	perfect := make([]byte, 5)
	for i := 0; i < 5; i++ {
		w := make(map[byte]int)
		for _, c := range candidates {
			w[c[i]]++
		}
		m := 0
		for k, v := range w {
			if v >= m {
				perfect[i] = k
				m = v
			}
		}
	}
	for i, c := range candidates {
		for j := 0; j < 5; j++ {
			if c[j] == perfect[j] {
				score[i]++
			}
		}
		// consider unique letters
		for j := 0; j < 4; j++ {
			for k := j + 1; k < 5; k++ {
				if c[j] == c[k] {
					score[i]--
				}
			}
		}
	}
	m := 1
	guess := ""
	for i, c := range candidates {
		//fmt.Println(c, score[i])
		if score[i] > m {
			m = score[i]
			guess = c
		}
	}
	return guess
}

type Config struct {
	ShowCandidateCount bool
	ShowBestGuess      bool
}

var config Config

func main() {
	flag.BoolVar(&config.ShowCandidateCount, "c", false, "Show count of possible candidates")
	flag.BoolVar(&config.ShowBestGuess, "g", false, "Show best guess")

	flag.Parse()

	validWord := regexp.MustCompile("^[a-z]{5}$")

	var guesses []Guess

	for _, a := range flag.Args() {
		p := strings.Split(a, ":")
		if len(p) != 2 {
			fmt.Printf("argument %#v is not the right length\n", a)
			os.Exit(1)
		}
		if len(p[0]) != len(p[1]) {
			fmt.Printf("Guess and result are not of the same length: %s: %d != %d\n", a, len(p[0]), len(p[1]))
			os.Exit(1)
		}
		guesses = append(guesses, Guess{Word: p[0], Result: p[1]})
	}

	yellows := getYellows(guesses)
	//fmt.Printf("Yellows: %#v\n", strings.Join(yellows, ""))

	regexPattern := buildRegex(guesses)
	//fmt.Printf("Pattern: %#v\n", regexPattern)
	validGuess := regexp.MustCompile(regexPattern)

	file, err := words.Open("wordles.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	candidates := []string{}
	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		t := scanner.Text()
		if !validWord.MatchString(t) {
			continue
		}
		if !validGuess.MatchString(t) {
			continue
		}
		foundYellow := true
		for _, y := range yellows {
			if strings.Index(t, y) == -1 {
				foundYellow = false
				break
			}
		}
		if foundYellow {
			//fmt.Printf("Adding candidates: %s\n", t)
			candidates = append(candidates, t)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	bestGuess := getBestGuess(candidates)

	output := false
	if config.ShowCandidateCount {
		fmt.Printf("Possible Candidates: %d\n", len(candidates))
		output = true
	}
	if config.ShowBestGuess {
		fmt.Printf("Best Guess: %s\n", bestGuess)
		output = true
	}
	if !output {
		if len(candidates) > 0 {
			fmt.Println("Hmm, yes, there's a wordle")
		} else {
			fmt.Println("Nope, no wordle here")
		}
	}

}
