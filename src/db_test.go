package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestInsertCard(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	tables := []struct {
		name     string
		cardItem CardItem
		result   sql.Result
		expected error
	}{
		{
			name: "success",
			cardItem: CardItem{
				Question: "What is the capital of France?",
				Answer:   "Paris",
			},
			result:   sqlmock.NewResult(1, 1),
			expected: nil,
		},
		{
			name: "failure",
			cardItem: CardItem{
				Question: "What is the capital of France?",
				Answer:   "Paris",
			},
			result:   nil,
			expected: errors.New("Error in preparing statement"),
		},
	}
	for _, table := range tables {
		t.Run(table.name, func(t *testing.T) {
			mock.ExpectPrepare("insert into cards").
				ExpectExec().
				WithArgs(table.cardItem.Question, table.cardItem.Answer).
				WillReturnResult(table.result)
			err := insertCard(table.cardItem, db)
			assert.Equal(t, table.expected, err)
			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestInsertMCQ(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	mcq := MCQItem{
		Question: "What is the capital of France?",
		Answer:   "Paris",
		Options:  []string{"Paris", "London", "Berlin", "Rome"},
	}
	joined := strings.Join(mcq.Options[:], ",")
	mock.ExpectPrepare("insert into cards").
		ExpectExec().
		WithArgs(mcq.Question, mcq.Answer, joined).
		WillReturnResult(sqlmock.NewResult(1, 1))
	err = insertMCQ(mcq, db)
	assert.NoError(t, err)
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestInsertMCQFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	q := MCQItem{
		Question: "What is the capital of France?",
		Answer:   "Paris",
		Options:  []string{"Paris", "London", "Berlin", "Rome"},
	}

	mock.ExpectPrepare("insert into cards (question, answer, options) values (?,?,?)").
		ExpectExec().
		WithArgs(q.Question, q.Answer, "Paris, London, Berlin, Rome").
		WillReturnError(fmt.Errorf("something went wrong"))

	err = insertMCQ(q, db)
	require.Error(t, err)
	assert.EqualError(t, err, "Error in preparing statement")
}
