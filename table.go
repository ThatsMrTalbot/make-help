package main

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

type Color string

const (
	ColorNone      Color = "\033[0m"
	ColorGreen     Color = "\033[0;32m"
	ColorLightBlue Color = "\033[0;94m"
)

const zeroWidthSpace = "\u200B"

type ColorTableWriter struct {
	rows       int
	currentRow []cell
	tw         *tabwriter.Writer
}

func NewColorTableWriter(w io.Writer, rows int) ColorTableWriter {
	return ColorTableWriter{
		rows: rows,
		tw:   tabwriter.NewWriter(w, 0, 4, 4, ' ', 0),
	}
}

type cell struct {
	color Color
	lines []string
}

func (ctw *ColorTableWriter) AddCell(color Color, contents string) {
	ctw.currentRow = append(ctw.currentRow, cell{
		color: color,
		lines: strings.Split(contents, "\n"),
	})
}

func (ctw *ColorTableWriter) FlushRow() {
	line := 0
	for {
		done := true

		// If we were not provided enough rows, then print some leading empty
		// cells. This is useful for this specific implementation.
		if delta := ctw.rows - len(ctw.currentRow); delta > 0 {
			fmt.Fprint(ctw.tw, strings.Repeat(strings.Repeat(zeroWidthSpace, 11)+"\t", delta))
		}

		for i, cell := range ctw.currentRow {
			// If this is not the first cell, print the tab separator
			if i != 0 {
				fmt.Fprint(ctw.tw, "\t")
			}

			// Print the color, then the line
			fmt.Fprint(ctw.tw, cell.color)
			if len(cell.lines) > line {
				fmt.Fprint(ctw.tw, cell.lines[line])
			}

			// We are done when there are no further lines
			done = done && len(cell.lines) <= line+1

			// Reset the color
			fmt.Fprint(ctw.tw, ColorNone)

			// Pad with spaces to ensure the color codes are all the same width
			if delta := 7 - len(cell.color); delta > 0 {
				fmt.Fprint(ctw.tw, strings.Repeat(zeroWidthSpace, delta))
			}
		}

		fmt.Fprint(ctw.tw, "\n")

		if done {
			break
		}

		line++
	}

	ctw.currentRow = ctw.currentRow[:0]
}

func (ctw *ColorTableWriter) FlushTable() {
	ctw.tw.Flush()
}
