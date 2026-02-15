// Package table provides a simple ASCII table formatter with customizable borders and colors.
package table

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/ipanardian/lu-hut/pkg/helper"
)

const (
	StyleSingle = iota
	StyleDouble
	StyleBold
)

type Table struct {
	data         [][]string
	borderStyle  int
	headerStyle  int
	headerColor  *color.Color
	borderColor  *color.Color
	columnWidths []int
	totalWidth   int
}

type borderChars struct {
	horizontal  string
	vertical    string
	topLeft     string
	topRight    string
	bottomLeft  string
	bottomRight string
	middle      string
	topTee      string
	bottomTee   string
	leftTee     string
	rightTee    string
	cross       string
}

func NewTable(data [][]string) *Table {
	t := &Table{
		data:        data,
		borderStyle: StyleSingle,
		headerStyle: StyleBold,
	}
	t.calculateColumnWidths()
	return t
}

func NewTableWithWidths(data [][]string, widths []int) *Table {
	t := &Table{
		data:         data,
		borderStyle:  StyleSingle,
		headerStyle:  StyleBold,
		columnWidths: widths,
	}
	t.calculateTotalWidth()
	return t
}

func (t *Table) SetBorderStyle(style int) {
	t.borderStyle = style
}

func (t *Table) SetHeaderStyle(style int) {
	t.headerStyle = style
}

func (t *Table) SetHeaderColor(c *color.Color) {
	t.headerColor = c
}

func (t *Table) SetBorderColor(c *color.Color) {
	t.borderColor = c
}

func (t *Table) getBorderChars() borderChars {
	switch t.borderStyle {
	case StyleDouble:
		return borderChars{
			horizontal:  "═",
			vertical:    "║",
			topLeft:     "╔",
			topRight:    "╗",
			bottomLeft:  "╚",
			bottomRight: "╝",
			middle:      "╬",
			topTee:      "╦",
			bottomTee:   "╩",
			leftTee:     "╠",
			rightTee:    "╣",
			cross:       "╬",
		}
	case StyleBold:
		return borderChars{
			horizontal:  "━",
			vertical:    "┃",
			topLeft:     "┏",
			topRight:    "┓",
			bottomLeft:  "┗",
			bottomRight: "┛",
			middle:      "╋",
			topTee:      "┳",
			bottomTee:   "┻",
			leftTee:     "┣",
			rightTee:    "┫",
			cross:       "╋",
		}
	default:
		return borderChars{
			horizontal:  "─",
			vertical:    "│",
			topLeft:     "┌",
			topRight:    "┐",
			bottomLeft:  "└",
			bottomRight: "┘",
			middle:      "┼",
			topTee:      "┬",
			bottomTee:   "┴",
			leftTee:     "├",
			rightTee:    "┤",
			cross:       "┼",
		}
	}
}

func (t *Table) calculateTotalWidth() {
	for _, width := range t.columnWidths {
		t.totalWidth += width
	}
	t.totalWidth += (len(t.columnWidths)-1)*3 + 2
}

func (t *Table) Print() {
	if len(t.data) == 0 {
		return
	}

	bc := t.getBorderChars()

	t.printTopBorder(bc)
	t.printRow(0, bc, true)

	if len(t.data) > 1 {
		t.printSeparator(bc)
		for i := 1; i < len(t.data); i++ {
			t.printRow(i, bc, false)
		}
	}

	t.printBottomBorder(bc)
}

func (t *Table) printTopBorder(bc borderChars) {
	var line strings.Builder
	line.WriteString(bc.topLeft)
	for i := range t.columnWidths {
		line.WriteString(strings.Repeat(bc.horizontal, t.columnWidths[i]+2))
		if i < len(t.columnWidths)-1 {
			line.WriteString(bc.topTee)
		}
	}
	line.WriteString(bc.topRight)
	t.printColored(line.String(), t.borderColor)
}

func (t *Table) printBottomBorder(bc borderChars) {
	var line strings.Builder
	line.WriteString(bc.bottomLeft)
	for i := range t.columnWidths {
		line.WriteString(strings.Repeat(bc.horizontal, t.columnWidths[i]+2))
		if i < len(t.columnWidths)-1 {
			line.WriteString(bc.bottomTee)
		}
	}
	line.WriteString(bc.bottomRight)
	t.printColored(line.String(), t.borderColor)
}

func (t *Table) printSeparator(bc borderChars) {
	var line strings.Builder
	line.WriteString(bc.leftTee)
	for i := range t.columnWidths {
		line.WriteString(strings.Repeat(bc.horizontal, t.columnWidths[i]+2))
		if i < len(t.columnWidths)-1 {
			line.WriteString(bc.middle)
		}
	}
	line.WriteString(bc.rightTee)
	t.printColored(line.String(), t.borderColor)
}

func (t *Table) calculateColumnWidths() {
	if len(t.data) == 0 {
		return
	}

	t.columnWidths = make([]int, len(t.data[0]))

	for _, row := range t.data {
		for j, cell := range row {
			width := utf8.RuneCountInString(cell)
			if width > t.columnWidths[j] {
				t.columnWidths[j] = width
			}
		}
	}

	t.calculateTotalWidth()
}

func (t *Table) printColored(text string, c *color.Color) {
	if c != nil {
		c.Println(text)
	} else {
		fmt.Println(text)
	}
}

func (t *Table) printRow(rowIndex int, bc borderChars, isHeader bool) {
	row := t.data[rowIndex]

	if t.borderColor != nil {
		t.borderColor.Print(bc.vertical)
	} else {
		fmt.Print(bc.vertical)
	}

	for i := 0; i < len(t.columnWidths); i++ {
		cell := ""
		if i < len(row) {
			cell = row[i]
		}
		cellWidth := utf8.RuneCountInString(helper.StripANSI(cell))
		maxWidth := t.columnWidths[i]

		var cellContent string
		if cellWidth > maxWidth {
			truncated := truncateString(cell, maxWidth)
			truncatedWidth := min(utf8.RuneCountInString(helper.StripANSI(truncated)), maxWidth)
			padding := max(maxWidth-truncatedWidth, 0)
			cellContent = " " + truncated + strings.Repeat(" ", padding) + " "
		} else {
			padding := maxWidth - cellWidth
			rightPad := padding
			leftPad := 0
			if isHeader {
				leftPad = padding / 2
				rightPad = padding - leftPad
			}
			cellContent = " " + strings.Repeat(" ", leftPad) + cell + strings.Repeat(" ", rightPad) + " "
		}

		if isHeader {
			fmt.Print(t.printColoredReturn(cellContent, t.headerColor))
		} else {
			fmt.Print(cellContent)
		}

		if i < len(t.columnWidths)-1 {
			if t.borderColor != nil {
				t.borderColor.Print(bc.vertical)
			} else {
				fmt.Print(bc.vertical)
			}
		}
	}

	if t.borderColor != nil {
		t.borderColor.Println(bc.vertical)
	} else {
		fmt.Println(bc.vertical)
	}
}

func truncateString(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}

	plain := helper.StripANSI(s)
	runes := []rune(plain)
	if len(runes) <= maxLen {
		return s
	}

	if maxLen <= 3 {
		return string(runes[:maxLen])
	}

	if maxLen > 7 {
		keepStart := (maxLen - 3) / 2
		keepEnd := (maxLen - 3) - keepStart
		return string(runes[:keepStart]) + "..." + string(runes[len(runes)-keepEnd:])
	}

	return string(runes[:maxLen-3]) + "..."
}

func (t *Table) printColoredReturn(text string, c *color.Color) string {
	if c != nil {
		return c.SprintFunc()(text)
	}
	return text
}
