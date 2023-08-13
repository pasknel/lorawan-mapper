package util

import (
	"os"

	"github.com/jedib0t/go-pretty/table"
)

func CreateTable(header table.Row, rows []table.Row) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(header)
	t.AppendRows(rows)
	t.SetStyle(table.StyleLight)
	t.Render()
}
