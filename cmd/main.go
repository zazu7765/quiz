package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)
type BaseItem struct{
	id int
}
type ItemInterface interface {
}
type CardItem struct {
	ItemInterface
	BaseItem
	Question string
	Answer   string
}
type MCQItem struct {
	ItemInterface
	BaseItem
	Question string
	Options  []string
	Answer   string
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
	files, err := os.ReadDir(ROOT)
	for _, file := range files {
		// fmt.Println(file.Name())
		if filepath.Ext(file.Name()) == ".db" {
			db, quiz, err := openDeck(file.Name())
			if err != nil {
				fmt.Println(err)
			} else {
				items, err := parseDeck(db, quiz)
				if err != nil {
					fmt.Println(err)
				} else {
					for _, i := range items {
						switch item := i.(type) {
						case CardItem:
							fmt.Println(item.Question)
						case MCQItem:
							fmt.Println(item.Question)
						default:
							fmt.Printf("type: %T\n", item)
						}
					}
				}
			}
			defer db.Close()
		}
	}
	// TODO: continue flow here
	// test1 := CardItem{
	// 	Question: "Question 1",
	// 	Answer:   "Answer 1",
	// }
	// test2 := MCQItem{
	// 	Question: "MCQ Question 1",
	// 	Answer:   "Option 1",
	// 	Options:  []string{"Option 1", "Option 2", "Option 3"},
	// }
	// card_db, _, err := openDeck("CRD_questions.db")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// err = insertCard(test1, card_db)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// mcq_db, _, err := openDeck("MCQ_flashcards.db")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// err = insertMCQ(test2, mcq_db)
	// if err != nil {
	// 	fmt.Println(err)
	// }
}
