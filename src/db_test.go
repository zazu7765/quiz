package main

import "testing"

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
