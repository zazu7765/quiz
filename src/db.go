package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

type Quiz int64

const (
	Undefined Quiz = iota
	Card
	MCQ
)

func (q Quiz) String() string {
	switch q {
	case Card:
		return "CRD"
	case MCQ:
		return "MCQ"
	}
	return "unknown"
}

func createDeck(n string, t Quiz) error {
	dbName := ROOT + "/" + t.String() + "_" + n + ".db"
	if _, err := os.Stat(dbName); os.IsExist(err) {
		return errors.New("Quiz already exists")
	}
	db, err := sql.Open("sqlite", dbName)
	if err != nil {
		return err
	}
	if t == MCQ {
		db.Exec("CREATE TABLE cards (id INT primary key, question TEXT not null, options TEXT not null, answer TEXT not null);")
	} else if t == Card {
		db.Exec("CREATE TABLE cards (id INT primary key, question TEXT not null, answer TEXT not null);")
	}
	defer db.Close()
	return nil
}

func parseDeck(n string) (*sql.DB, error) {
	if n[:4] != "CRD_" && n[:4] != "MCQ_" {
		return nil, errors.New("Invalid Database Type! Must be MCQ or CRD")
	}
	file := ROOT + "/" + n
	db, err := sql.Open("sqlite", file)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("Select * from cards")
	if err != nil {
		db.Close()
		return nil, errors.New(fmt.Sprintf("Illegal Database: %s", file))
	}
	return db, nil
}

func insertItem(n CardItem /*, db *sql.DB*/) error {
	question := n.Question
	answer := n.Answer
	if n.Type == MCQ {
		answer = strings.Join(n.Options[:], ",")
	}
	fmt.Println(question)
	fmt.Println(answer)
	return nil
}
