package main

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestQuizEnums(t *testing.T) {
	tables := []struct {
		q Quiz
		s string
	}{
		{Card, "CRD"},
		{MCQ, "MCQ"},
		{Undefined, "Unknown"},
	}
	for _, table := range tables {
		str := table.q.String()
		if str != table.s {
			t.Errorf("Enum of %T was incorrect, got %s instead of %s", table.q, str, table.s)
		}
	}
}

//	func TestParseMCQDeck(t *testing.T) {
//		db, mock, err := sqlmock.New()
//		if err != nil {
//			t.Errorf("Error opening stub database: %s", err)
//		}
//		defer db.Close()
//		rows := sqlmock.NewRows([]string{"id", "question", "options", "answer"}).AddRow(1, "test question 1", "option1, option2, option3", "option2")
//		mock.ExpectQuery("select question, options, answer from cards").WillReturnRows(rows)
//		_, err = parseDeck(db, MCQ)
//		if err != nil {
//			t.Errorf("Error running test: %s", err)
//		}
//
//		if err := mock.ExpectationsWereMet(); err != nil {
//			t.Errorf("Unfulfilled Expectations: %s", err)
//		}
//	}
func TestParseDeck(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	tests := []struct {
		q     Quiz
		rows  *sqlmock.Rows
		items []ItemInterface
		err   error
	}{
		{
			q: Card,
			rows: sqlmock.NewRows([]string{"question", "answer"}).
				AddRow("What is the capital of France?", "Paris").
				AddRow("What is the tallest mountain in the world?", "Mount Everest"),
			items: []ItemInterface{
				CardItem{Question: "What is the capital of France?", Answer: "Paris"},
				CardItem{Question: "What is the tallest mountain in the world?", Answer: "Mount Everest"},
			},
			err: nil,
		},
		{
			q: MCQ,
			rows: sqlmock.NewRows([]string{"question", "options", "answer"}).
				AddRow("What is the capital of France?", "Paris,London,Berlin,Rome", "Paris").
				AddRow("What is the tallest mountain in the world?", "Mount Kilimanjaro,Mount Everest,K2,Mont Blanc", "Mount Everest"),
			items: []ItemInterface{
				MCQItem{Question: "What is the capital of France?", Options: []string{"Paris", "London", "Berlin", "Rome"}, Answer: "Paris"},
				MCQItem{Question: "What is the tallest mountain in the world?", Options: []string{"Mount Kilimanjaro", "Mount Everest", "K2", "Mont Blanc"}, Answer: "Mount Everest"},
			},
			err: nil,
		},
		{
			q: Card,
			rows: sqlmock.NewRows([]string{"question", "answer"}).
				RowError(1, sql.ErrNoRows),
			items: nil,
			err:   sql.ErrNoRows,
		},
	}

	for _, test := range tests {
		t.Run(test.q.String(), func(t *testing.T) {
			mock.ExpectQuery("select .* from cards").WillReturnRows(test.rows)

			items, err := parseDeck(db, test.q)

			assert.Equal(t, test.err, err)
			assert.Equal(t, test.items, items)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
