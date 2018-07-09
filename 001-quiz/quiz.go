package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	// "runtime"
	"strings"
	"sync"
	"time"
)

var (
	wg           sync.WaitGroup
)

type quiz struct {
	challenge, response string
}

func main() {
	qs, err := readCSV("problems.csv")
	if err != nil {
		fmt.Println(err)
	} else {
		play(qs)
	}
}

func play(qs []quiz) {
	totalQuestions := len(qs)
	responses := make(map[int]string, totalQuestions)

	respondTo := make(chan string)

	// block until user presses enter
	fmt.Println("Press [Enter] to start test.")
	bufio.NewScanner(os.Stdout).Scan()

	wg.Add(1)
	timeUp := time.After(time.Second * time.Duration(2))
	go func() {
	label:
		for i := 0; i < totalQuestions; i++ {
			//index := randPool[i]
			go askQuestion(os.Stdout, os.Stdin, qs[i].challenge, respondTo)
			select {
			case <-timeUp:
				fmt.Fprintln(os.Stderr, "\nTime up!")
				break label
			case ans, ok := <-respondTo:
				if ok {
					responses[i] = ans
				} else {
					break label
				}
			}
		}
		wg.Done()
	}()
	wg.Wait()

	// reader := bufio.NewReader(os.Stdin)
	numCorrect := 0
	// for i := 0; i < len(qs); i++ {
	// 	fmt.Println(fmt.Sprintf("Problem #%v: %v = ", i+1, qs[i].challenge))
	// 	text, _ := reader.ReadString('\n')
	// 	if runtime.GOOS == "windows" {
	// 		text = strings.TrimRight(text, "\r\n")
	// 	  } else {
	// 		text = strings.TrimRight(text, "\n")
	// 	  }
	// 	if (text == qs[i].response) {
	// 		numCorrect++
	// 	}
	// }
	for i := 0; i < len(responses); i++ {
		if checkAnswer(responses[i], qs[i].response) {
			numCorrect++
		}
	}
	fmt.Println(fmt.Sprintf("You scored %v out of %v.", numCorrect, totalQuestions))
}

func askQuestion(w io.Writer, r io.Reader, question string, replyTo chan string) {
	reader := bufio.NewReader(r)
	fmt.Fprintln(w, "Question: "+question)
	fmt.Fprint(w, "Answer: ")
	answer, err := reader.ReadString('\n')
	if err != nil {
		close(replyTo)
		if err == io.EOF {
			return
		}
		log.Fatalln(err)
	}
	replyTo <- answer
}

func checkAnswer(ans string, expected string) bool {
	if strings.EqualFold(strings.TrimSpace(ans), strings.TrimSpace(expected)) {
		return true
	}
	return false
}

func readCSV(filename string) ([]quiz, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("couldn't open file: %v", err)
	}
	defer f.Close() // nolint

	r := csv.NewReader(f)
	out := []quiz{}

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error while reading CSV record: %v", err)
		}
		if len(record) != 2 {
			return nil, fmt.Errorf("unexpected number of fields for record: %v", record)
		}

		out = append(out, quiz{record[0], record[1]})
	}

	return out, nil
}
