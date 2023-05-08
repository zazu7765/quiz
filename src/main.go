package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type CardItem struct {
	Question string
	Options  []string
	Answer   string
	Type     Quiz
}

var ROOT string

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	// define root directory for all operations
	ROOT = home + "/.quiz"
	if _, err := os.Stat(ROOT); os.IsNotExist(err) {
		os.Mkdir(ROOT, 0750)
	}
	// createDeck("banana", Card)
	files, err := os.ReadDir(ROOT)
	for _, file := range files {
		// fmt.Println(file.Name())
		if filepath.Ext(file.Name()) == ".db" {
			db, err := parseDeck(file.Name())
			if err != nil {
				fmt.Println(err)
			} else {
				defer db.Close()
			}
		}
	}
	// TODO: continue flow here
	test1 := CardItem{
		Question: "Question 1",
		Options:  []string{},
		Answer:   "Answer 1",
		Type:     Card,
	}
	test2 := CardItem{
		Question: "MCQ Question 1",
		Options:  []string{"Option 1", "Option 2", "Option 3"},
		Answer:   "Option 1",
		Type:     MCQ,
	}
	insertItem(test1)
	insertItem(test2)
}
