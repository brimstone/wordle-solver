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
	"strconv"
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
			if g.ResultA()[i] == "_" {
				continue
			}
			log.Fatal(fmt.Sprintf("Marker: \"%s\" is unknown", g.ResultA()[i]))

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

func solve(guesses []Guess) (string, int) {
	validWord := regexp.MustCompile("^[a-z]{5}$")

	yellows := getYellows(guesses)
	regexPattern := buildRegex(guesses)
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
	return bestGuess, len(candidates)
}

var config Config

func main() {
	flag.BoolVar(&config.ShowCandidateCount, "c", false, "Show count of possible candidates")
	flag.BoolVar(&config.ShowBestGuess, "g", false, "Show best guess")

	flag.Parse()

	var guesses [][]Guess
	var solved []bool

	for _, a := range flag.Args() {
		p := strings.Split(a, ":")
		if len(p) < 2 {
			fmt.Printf("argument %#v is not the right length\n", a)
			os.Exit(1)
		}
		for i := 1; i < len(p); i++ {
			if len(guesses) < i {
				guesses = append(guesses, []Guess{})
				solved = append(solved, false)
			}
			if len(p[0]) != len(p[i]) {
				fmt.Printf("Guess and result are not of the same length: %s:%s: %d != %d\n", p[0], p[i], len(p[0]), len(p[i]))
				os.Exit(1)
			}
			guesses[i-1] = append(guesses[i-1], Guess{Word: p[0], Result: strings.ToLower(p[i])})
			if strings.ToLower(p[i]) == "ggggg" {
				solved[i-1] = true
			}
			//fmt.Printf("Guesses: %s %#v\n", p[i], guesses)
		}
	}

	var bestguesses []string
	var candidates []int
	for g := range guesses {
		b, c := solve(guesses[g])
		bestguesses = append(bestguesses, b)
		candidates = append(candidates, c)
	}

	output := false
	if config.ShowCandidateCount {
		o := ""
		for _, c := range candidates {
			o += ", " + strconv.FormatInt(int64(c), 10)
		}
		fmt.Printf("Possible Candidates: %s\n", o[2:])
		output = true
	}
	if config.ShowBestGuess {
		n := ""
		if len(candidates) > 1 {
			n = "es"
		}
		fmt.Printf("Best Guess%s: %s\n", n, strings.Join(bestguesses, ", "))
		output = true
		if len(candidates) > 1 {
			m := 0
			for i, s := range solved {
				m = i
				if !s {
					break
				}
			}
			for i, c := range candidates {
				if solved[i] {
					continue
				}
				if candidates[m] > c {
					m = i
				}
			}
			fmt.Printf("Bestest Guess: %s\n", bestguesses[m])
		}
	}
	if !output {
		if candidates[0] > 0 {
			fmt.Println("Hmm, yes, there's a wordle")
		} else {
			fmt.Println("Nope, no wordle here")
		}
	}

}
