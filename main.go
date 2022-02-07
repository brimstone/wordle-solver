package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
)

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

func main() {
	validWord := regexp.MustCompile("^[a-z]{5}$")

	guesses := []Guess{
		Guess{Word: "salet", Result: "_ygy_"},
		Guess{Word: "abide", Result: "y___y"},
		Guess{Word: "relay", Result: "_ggy_"},
	}

	yellows := getYellows(guesses)
	fmt.Printf("Yellows: %#v\n", strings.Join(yellows, ""))

	regexPattern := buildRegex(guesses)
	fmt.Printf("Pattern: %#v\n", regexPattern)
	validGuess := regexp.MustCompile(regexPattern)

	file, err := os.Open("/usr/share/dict/words")
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
			fmt.Printf("Adding candidates: %s\n", t)
			candidates = append(candidates, t)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	bestGuess := getBestGuess(candidates)

	fmt.Printf("Try: %s\n", bestGuess)

}
