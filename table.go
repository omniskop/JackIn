package main

import (
	"fmt"
	"strings"
)

// Table provides a simple way to create a printable table
type Table struct {
	rows        [][]string
	columnSizes []int
}

func (t *Table) Add(row ...any) {
	strRow := make([]string, len(row))
	for i, col := range row {
		strRow[i] = fmt.Sprint(col)
	}
	t.rows = append(t.rows, strRow)
	if len(row) > len(t.columnSizes) {
		t.columnSizes = append(t.columnSizes, make([]int, len(row)-len(t.columnSizes))...)
	}
	for i, col := range strRow {
		if len(col) > t.columnSizes[i] {
			t.columnSizes[i] = len(col)
		}
	}
}

func (t *Table) String() string {
	var b strings.Builder
	for _, row := range t.rows {
		for i, col := range row {
			b.WriteString(col)
			for j := len(col); j < t.columnSizes[i]; j++ {
				b.WriteByte(' ')
			}
			b.WriteByte(' ')
		}
		b.WriteByte('\n')
	}
	return b.String()
}
