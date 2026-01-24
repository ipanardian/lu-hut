// Package renderer provides table rendering and formatting functionality.
package renderer

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/ipanardian/lu-hut/internal/config"
	"github.com/ipanardian/lu-hut/internal/model"
	"github.com/ipanardian/lu-hut/internal/table"
)

type Table struct {
	config config.Config
}

func NewTable(cfg config.Config) *Table {
	return &Table{config: cfg}
}

func (r *Table) Render(files []model.FileEntry, now time.Time) {
	if len(files) == 0 {
		return
	}

	terminalWidth := max(getTerminalWidth(), 40)

	mins, maxs := r.columnConstraints()
	nameWidth := 0
	if len(maxs) > 0 {
		nameWidth = maxs[0]
	}

	data := r.buildTableData(files, now, nameWidth)
	displayWidths := calculateDisplayWidths(data)

	for i := range displayWidths {
		if i < len(mins) && mins[i] > 0 && displayWidths[i] < mins[i] {
			displayWidths[i] = mins[i]
		}
		if i < len(maxs) && maxs[i] > 0 && displayWidths[i] > maxs[i] {
			displayWidths[i] = maxs[i]
		}
	}

	minContentWidth := 0
	for i := range displayWidths {
		minContentWidth += lookupMin(mins, i, 4)
	}
	minBorderWidth := (len(displayWidths)-1)*3 + 2
	if terminalWidth < minContentWidth+minBorderWidth {
		fmt.Println("Terminal is too small to display the table. Please widen your terminal window.")
		return
	}

	totalContentWidth := 0
	for _, w := range displayWidths {
		totalContentWidth += w
	}
	borderWidth := (len(displayWidths)-1)*3 + 2
	totalWidth := totalContentWidth + borderWidth

	if totalWidth > terminalWidth {
		r.shrinkColumns(displayWidths, mins, totalWidth-terminalWidth)
	}

	tbl := table.NewTableWithWidths(data, displayWidths)
	tbl.SetBorderStyle(0)
	tbl.SetHeaderStyle(1)
	tbl.SetHeaderColor(color.New(color.FgWhite, color.Bold))
	tbl.SetBorderColor(color.New(color.FgGreen))
	tbl.Print()
}

func (r *Table) buildTableData(files []model.FileEntry, now time.Time, nameWidth int) [][]string {
	headers := []string{"Name", "Size", "Modified", "Perms"}
	if r.config.ShowGit {
		headers = append(headers, "Git")
	}
	if r.config.ShowUser {
		headers = append(headers, "User", "Group")
	}

	data := make([][]string, len(files)+1)
	data[0] = headers

	for i, file := range files {
		row := []string{
			formatName(file, nameWidth),
			formatSize(file.Size, file.IsDir),
			formatModified(file.ModTime, now, r.config.ShowExactTime),
			formatPermissions(file.Mode, r.config.ShowOctal),
		}
		if r.config.ShowGit {
			row = append(row, formatGitStatus(file.GitStatus))
		}
		if r.config.ShowUser {
			row = append(row, file.Author, file.Group)
		}
		data[i+1] = row
	}

	return data
}

func (r *Table) columnConstraints() ([]int, []int) {
	mins := []int{15, 6, 10, 10}
	maxs := []int{50, 10, 15, 12}
	if r.config.ShowExactTime {
		mins[2] = 16
		maxs[2] = 17
	}
	if r.config.ShowGit {
		mins = append(mins, 6)
		maxs = append(maxs, 12)
	}
	if r.config.ShowUser {
		mins = append(mins, 6, 6)
		maxs = append(maxs, 12, 12)
	}
	return mins, maxs
}

func (r *Table) shrinkColumns(displayWidths, mins []int, excess int) {
	totalShrinkable := 0
	for i, w := range displayWidths {
		if i != 1 && i != 3 {
			minWidth := lookupMin(mins, i, 4)
			if w-minWidth > 0 {
				totalShrinkable += w - minWidth
			}
		}
	}

	for i := range displayWidths {
		if i != 1 && i != 3 {
			minWidth := lookupMin(mins, i, 4)
			shrinkable := displayWidths[i] - minWidth
			if shrinkable > 0 && totalShrinkable > 0 {
				shrinkAmount := (shrinkable * excess) / totalShrinkable
				shrinkAmount = min(shrinkAmount, shrinkable)
				displayWidths[i] -= shrinkAmount
				displayWidths[i] = max(displayWidths[i], minWidth)
				excess -= shrinkAmount
				totalShrinkable -= shrinkAmount
			}
		}
	}
}
