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
	return "Unknown"
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
		db.Exec("CREATE TABLE cards (question TEXT not null, options TEXT not null, answer TEXT not null);")
	} else if t == Card {
		db.Exec("CREATE TABLE cards (question TEXT not null, answer TEXT not null);")
	}
	defer db.Close()
	return nil
}

func openDeck(n string) (*sql.DB, Quiz, error) {
	var q Quiz
	if n[:4] != "CRD_" && n[:4] != "MCQ_" {
		return nil, Undefined, errors.New("Invalid Database Type! Must be MCQ or CRD")
	}
	if n[:3] == "CRD" {
		q = Card
	} else {
		q = MCQ
	}
	file := ROOT + "/" + n
	db, err := sql.Open("sqlite", file)
	if err != nil {
		return nil, Undefined, err
	}
	_, err = db.Exec("Select * from cards")
	if err != nil {
		db.Close()
		return nil, Undefined, errors.New(fmt.Sprintf("Illegal Database: %s", file))
	}
	return db, q, nil
}

func parseDeck(db *sql.DB, q Quiz) ([]ItemInterface, error) {
	count := 0
	var deck []ItemInterface
	switch q {
	case Card:
		rows, err := db.Query("select rowid, question, answer from cards")
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var r CardItem
			err = rows.Scan(&r.BaseItem.id, &r.Question, &r.Answer)
			if err != nil {
				return nil, err
			}
			deck = append(deck, r)
			count++
		}
	case MCQ:
		rows, err := db.Query("select rowid, question, options, answer from cards")
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var r MCQItem
			var optionsString string
			err = rows.Scan(&r.BaseItem.id, &r.Question, &optionsString, &r.Answer)
			if err != nil {
				return nil, err
			}
			r.Options = strings.Split(optionsString, ",")
			deck = append(deck, r)
			count++
		}
	}
	if count < 1 {
		return deck, sql.ErrNoRows
	}
	return deck, nil
}

func insertCard(n CardItem, db *sql.DB) error {
	statement, err := db.Prepare("insert into cards (question, answer) values (?,?)")
	if err != nil {
		return err
	}
	question := n.Question
	answer := n.Answer
	_, err = statement.Exec(question, answer)
	if err != nil {
		return errors.New("Error in preparing statement")
	}
	return nil
}

func insertMCQ(n MCQItem, db *sql.DB) error {
	statement, err := db.Prepare("insert into cards (question, answer, options) values (?,?,?)")
	if err != nil {
		return errors.New("Error in preparing statement")
	}
	question := n.Question
	answer := n.Answer
	options := strings.Join(n.Options[:], ",")
	_, err = statement.Exec(question, answer, options)
	if err != nil {
		return errors.New("Error in preparing statement")
	}
	return nil
}

func retrieveCard(id int, db *sql.DB) (CardItem, error) {
	statement, err := db.Prepare("select * from cards where rowid=?")
	if err != nil {
		return CardItem{}, errors.New("Error preparing statement")
	}
	defer statement.Close()
	row := statement.QueryRow(id)
	var c CardItem
	err = row.Scan(&c.id, &c.Question, &c.Answer)
	if err != nil {
		if err == sql.ErrNoRows {
			return CardItem{}, errors.New("Card not found")
		}
		return CardItem{}, err
	}
	return c, nil
}
func retrieveMCQ(id int, db *sql.DB) (MCQItem, error) {
	return MCQItem{}, nil
}
