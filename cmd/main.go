package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

type BaseItem struct {
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

func findOrCreateDirectory(dir string) ([]fs.DirEntry, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	ROOT = filepath.Join(home, dir)
	if _, err := os.Stat(ROOT); errors.Is(err, fs.ErrNotExist) {
		err = os.Mkdir(ROOT, 0750)
		if err != nil {
			return nil, err
		}
	}
	files, err := os.ReadDir(ROOT)
	if err != nil {
		return files, err
	}
	return files, nil
}
func checkQuizFiles(files []fs.DirEntry)([]fs.DirEntry, error){
	var quizzes []fs.DirEntry
	for _, file := range files {
		// fmt.Println(file.Name())
		if filepath.Ext(file.Name()) == ".db" {
			db, quiz, err := openDeck(file.Name())
			if err != nil {
				return quizzes, err
			} else {
				_, err := parseDeck(db, quiz)
				if err != nil {
					return quizzes, err
				} else {
					quizzes = append(quizzes, file)
				}
			}
			defer db.Close()
		}
	}
	return quizzes, nil
}
func main() {
	dirFiles, err := findOrCreateDirectory(".quiz")
	if err != nil {
		log.Fatal("Error creating $HOME/.quiz folder or inaccessible filesystem permissions")
	}
	FILES, err := checkQuizFiles(dirFiles)
	if err!=nil{
		fmt.Println(err)
		log.Fatal("Error checking for available quizzes!")
	}
	fmt.Println(fmt.Sprintln("Files:", FILES))
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
	// err = createDeck("questions", Card)
	// if err !=nil{
	// 	fmt.Println(err)
	// }
	// card_db, _, err := openDeck("CRD_questions.db")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// err = insertCard(test1, card_db)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// err = createDeck("flashcards", MCQ)
	// if err !=nil{
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
