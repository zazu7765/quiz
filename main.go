package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
)

type Quiz int64
type CardItem struct {
	Question string
	Options  []string
	Answer   string
}

const (
	Undefined Quiz = iota
	Card
	MCQ
)

type Metadata struct {
	QuizType string `json:"QuizType"`
}

func (q Quiz) String() string {
	switch q {
	case Card:
		return "card"
	case MCQ:
		return "MCQ"
	}
	return "unknown"
}

var ROOT string

func createDeck(n string, t Quiz) error {
	dbName := ROOT + "/" + t.String() + "_" + n + ".db"
	if _, err := os.Stat(dbName); os.IsExist(err) {
		return errors.New("Quiz already exists")
	}
	db, err := sql.Open("sqlite", dbName)
	if err != nil {
		return err
	}
	db.Exec("CREATE TABLE cards (id INT primary key, question TEXT not null, answer TEXT not null);")
	defer db.Close()
	return nil
}

func parseDeck(n string) (*sql.DB, error) {
	file := ROOT + "/" + n
	db, err := sql.Open("sqlite", file)
	if err != nil {
		return nil, err
	}
	return db, nil
}
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
	createDeck("banana", Card)
	files, err := os.ReadDir(ROOT)
	for _, file := range files {
		fmt.Println(file.Name())
		if filepath.Ext(file.Name()) == ".db" {
			db, err := parseDeck(file.Name())
			defer db.Close()
			if err != nil {
				fmt.Println(err)
			}
			stmt, err := db.Prepare("insert into cards (question, answer) values (?,?)")
			if err != nil {
				log.Fatal(err)
			}
			defer stmt.Close()
			_, err = stmt.Exec("who is the current president of the united states", "joe biden")
			rows, err := db.Query("select question,answer from cards")
			if err != nil {
				log.Fatal(err)
			}
			defer rows.Close()
			for rows.Next() {
				var q, a string
				if err := rows.Scan(&q, &a); err != nil {
					log.Fatal(err)
				}
				fmt.Println(q, a)
			}
		}
	}
}
