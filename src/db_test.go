package main

import (
	"database/sql"
	"errors"
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
			rows: sqlmock.NewRows([]string{"id", "question", "answer"}).
				AddRow(0, "What is the capital of France?", "Paris").
				AddRow(1, "What is the tallest mountain in the world?", "Mount Everest"),
			items: []ItemInterface{
				CardItem{BaseItem: BaseItem{id: 0}, Question: "What is the capital of France?", Answer: "Paris"},
				CardItem{BaseItem: BaseItem{id: 1}, Question: "What is the tallest mountain in the world?", Answer: "Mount Everest"},
			},
			err: nil,
		},
		{
			q: MCQ,
			rows: sqlmock.NewRows([]string{"id", "question", "options", "answer"}).
				AddRow(0, "What is the capital of France?", "Paris,London,Berlin,Rome", "Paris").
				AddRow(1, "What is the tallest mountain in the world?", "Mount Kilimanjaro,Mount Everest,K2,Mont Blanc", "Mount Everest"),
			items: []ItemInterface{
				MCQItem{BaseItem: BaseItem{id: 0}, Question: "What is the capital of France?", Options: []string{"Paris", "London", "Berlin", "Rome"}, Answer: "Paris"},
				MCQItem{BaseItem: BaseItem{id: 1}, Question: "What is the tallest mountain in the world?", Options: []string{"Mount Kilimanjaro", "Mount Everest", "K2", "Mont Blanc"}, Answer: "Mount Everest"},
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

	tables := []struct {
		name     string
		mcqItem  MCQItem
		result   sql.Result
		expected error
	}{
		{
			name: "success",
			mcqItem: MCQItem{
				Question: "What is the capital of France?",
				Answer:   "Paris",
				Options:  []string{"Paris", "London", "Berlin", "Rome"},
			},
			result:   sqlmock.NewResult(1, 1),
			expected: nil,
		},
		{
			name: "failure",
			mcqItem: MCQItem{
				Question: "What is the capital of France?",
				Answer:   "Paris",
				Options:  []string{"Paris", "London", "Berlin", "Rome"},
			},
			result:   nil,
			expected: errors.New("Error in preparing statement"),
		},
	}
	for _, table := range tables {
		t.Run(table.name, func(t *testing.T) {
			joined := strings.Join(table.mcqItem.Options[:], ",")
			mock.ExpectPrepare("insert into cards").
				ExpectExec().
				WithArgs(table.mcqItem.Question, table.mcqItem.Answer, joined).
				WillReturnResult(table.result)
			err := insertMCQ(table.mcqItem, db)
			assert.Equal(t, table.expected, err)
			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestRetrieveCard(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	tables := []struct {
		name          string
		id            int
		eCard         CardItem
		eError        bool
		ePrepareError error
		eQueryError   error
	}{
		{name: "success",
			id: 0,
			eCard: CardItem{
				BaseItem: BaseItem{id: 0},
				Question: "What is the capital of France?",
				Answer:   "Paris",
			},
			eError:        false,
			ePrepareError: nil,
			eQueryError:   nil,
		},
		{name: "failure",
			id:       999,
			eCard:    CardItem{},
			eError: true,
			ePrepareError: nil,
			eQueryError: sql.ErrNoRows,
		},
	}
	for _, table := range tables {
		t.Run(table.name, func(t *testing.T) {
			mock.ExpectPrepare("select .* from cards where rowid=?").WillReturnError(table.ePrepareError)
			if !table.eError{
			rows := sqlmock.NewRows([]string{"rowid", "question", "answer"}).AddRow(table.eCard.id, table.eCard.Question, table.eCard.Answer)
			mock.ExpectQuery("select .* from cards where rowid=?").WithArgs(table.id).WillReturnRows(rows)
			}else{
				mock.ExpectQuery("select .* from cards where rowid=?").WithArgs(table.id).WillReturnError(table.eQueryError)
			}
			card, err := retrieveCard(table.id, db)
			if table.eError{
				require.Error(t, err)
				require.Empty(t, card, "Expected card to be empty")
			}else{
				require.NoError(t, err)
				require.Equal(t, table.eCard, card)
			}
			err = mock.ExpectationsWereMet()
			require.NoError(t, err)
		})
	}
}
