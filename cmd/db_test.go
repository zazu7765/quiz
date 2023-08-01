package main

import (
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
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
		assert.Equal(t, table.s, str)
	}
}

func TestParseDeck(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
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
	assert.NoError(t, err)
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
	assert.NoError(t, err)
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
	assert.NoError(t, err)
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
			id:            999,
			eCard:         CardItem{},
			eError:        true,
			ePrepareError: nil,
			eQueryError:   sql.ErrNoRows,
		},
	}
	for _, table := range tables {
		t.Run(table.name, func(t *testing.T) {
			mock.ExpectPrepare("select .* from cards where rowid=?").WillReturnError(table.ePrepareError)
			if !table.eError {
				rows := sqlmock.NewRows([]string{"rowid", "question", "answer"}).AddRow(table.eCard.id, table.eCard.Question, table.eCard.Answer)
				mock.ExpectQuery("select .* from cards where rowid=?").WithArgs(table.id).WillReturnRows(rows)
			} else {
				mock.ExpectQuery("select .* from cards where rowid=?").WithArgs(table.id).WillReturnError(table.eQueryError)
			}
			card, err := retrieveCard(table.id, db)
			if table.eError {
				assert.Error(t, err)
				assert.Empty(t, card, "Expected card to be empty")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, table.eCard, card)
			}
			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestRetrieveMCQ(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	tables := []struct {
		name          string
		id            int
		eMCQ          MCQItem
		eError        bool
		ePrepareError error
		eQueryError   error
	}{
		{name: "success",
			id: 0,
			eMCQ: MCQItem{
				BaseItem: BaseItem{id: 0},
				Question: "What is the capital of France?",
				Answer:   "Paris",
				Options:  []string{"Paris", "Berlin", "London", "Rome"},
			},
			eError:        false,
			ePrepareError: nil,
			eQueryError:   nil,
		},
		{name: "failure",
			id:            999,
			eMCQ:          MCQItem{},
			eError:        true,
			ePrepareError: nil,
			eQueryError:   sql.ErrNoRows,
		},
	}
	for _, table := range tables {
		t.Run(table.name, func(t *testing.T) {
			mock.ExpectPrepare("select .* from cards where rowid=?").WillReturnError(table.ePrepareError)
			if !table.eError {
				rows := sqlmock.NewRows([]string{"rowid", "question", "options", "answer"}).AddRow(table.eMCQ.id, table.eMCQ.Question, strings.Join(table.eMCQ.Options[:], ","), table.eMCQ.Answer)
				mock.ExpectQuery("select .* from cards where rowid=?").WithArgs(table.id).WillReturnRows(rows)
			} else {
				mock.ExpectQuery("select .* from cards where rowid=?").WithArgs(table.id).WillReturnError(table.eQueryError)
			}
			card, err := retrieveMCQ(table.id, db)
			if table.eError {
				assert.Error(t, err)
				assert.Empty(t, card, "Expected card to be empty")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, table.eMCQ, card)
			}
			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

func TestUpdateCard(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	tables := []struct {
		name      string
		toUpdate  CardItem
		eError    bool
		prepError error
		result    sql.Result
	}{
		{name: "Success",
			toUpdate: CardItem{
				BaseItem: BaseItem{id: 1},
				Question: "Updated Q",
				Answer:   "Updated A",
			},
			eError:    false,
			prepError: nil,
			result:    sqlmock.NewResult(1, 1),
		},
		{
			name: "Failure",
			toUpdate: CardItem{
				BaseItem: BaseItem{id: 999},
				Question: "Lorem ipsum",
				Answer:   "Dolor sit amet",
			},
			eError:    true,
			prepError: nil,
			result:    sqlmock.NewResult(0, 0),
		},
	}
	for _, table := range tables {
		t.Run(table.name, func(t *testing.T) {
			mock.ExpectPrepare("update cards set question=\\?, answer=\\? where rowid = \\?").WillReturnError(table.prepError)

			mock.ExpectExec("update cards set question=\\?, answer=\\? where rowid = \\?").WithArgs(table.toUpdate.Question, table.toUpdate.Answer, table.toUpdate.id).WillReturnResult(table.result)
			err = updateCard(table.toUpdate, db)
			if table.eError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUpdateMCQ(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()
	tables := []struct {
		name      string
		toUpdate  MCQItem
		eError    bool
		prepError error
		result    sql.Result
	}{
		{name: "Success",
			toUpdate: MCQItem{
				BaseItem: BaseItem{id: 1},
				Question: "Updated Q",
				Answer:   "Updated A",
				Options:  []string{"one", "two", "three"},
			},
			eError:    false,
			prepError: nil,
			result:    sqlmock.NewResult(1, 1),
		},
		{
			name: "Failure",
			toUpdate: MCQItem{
				BaseItem: BaseItem{id: 999},
				Question: "Lorem ipsum",
				Answer:   "Dolor sit amet",
				Options:  []string{"one", "two", "three"},
			},
			eError:    true,
			prepError: nil,
			result:    sqlmock.NewResult(0, 0),
		},
	}
	for _, table := range tables {
		t.Run(table.name, func(t *testing.T) {
			mock.ExpectPrepare("update cards set question=\\?, answer=\\?, options=\\? where rowid = \\?").WillReturnError(table.prepError)

			options := strings.Join(table.toUpdate.Options[:], ",")
			mock.ExpectExec("update cards set question=\\?, answer=\\?, options=\\? where rowid = \\?").WithArgs(table.toUpdate.Question, table.toUpdate.Answer, options, table.toUpdate.id).WillReturnResult(table.result)
			err = updateMCQ(table.toUpdate, db)
			if table.eError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
