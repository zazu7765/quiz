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
