package main

import (
	"database/sql"
	"errors"
	_ "modernc.org/sqlite"
	"os"
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
		return "Card"
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
