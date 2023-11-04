package excelreader

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

type ScheduleEntry struct {
	Date  time.Time
	Shift string
}

func FindScheduleEntries(reader io.ReadCloser, name string) ([]ScheduleEntry, error) {
	file, err := excelize.OpenReader(reader)

	if err != nil {
		return nil, err
	}

	sheets := file.GetSheetList()

	if len(sheets) != 1 {
		return nil, errors.New("ambiguous sheets (there isn't exactly 1 sheet, don't know which one to use). Make sure there is only 1 sheet in the file")
	}

	rows, err := file.GetRows(sheets[0])

	if err != nil {
		return nil, fmt.Errorf("error in iterating over rows: %w", err)
	}

	var datemapping map[int]time.Time

	for _, row := range rows {
		if len(row) == 0 {
			continue
		}

		if mapping, ok := findDateRow(row); ok {
			datemapping = mapping
		}

		if strings.HasPrefix(row[0], name) {
			if datemapping == nil {
				return nil, errors.New(fmt.Sprintf("found %s before knowing the dates", name))
			}

			// Construct a list of schedule entries
			entries := []ScheduleEntry{}

			for col, date := range datemapping {
				if col < len(row) {
					entries = append(entries, ScheduleEntry{
						Date:  date,
						Shift: row[col],
					})
				}
			}

			sort.Slice(entries, func(i, j int) bool {
				return entries[i].Date.Compare(entries[j].Date) == -1
			})

			return entries, nil
		}
	}

	return nil, errors.New("nothing found")
}

func findDateRow(row []string) (map[int]time.Time, bool) {
	datemap := make(map[int]time.Time)
	datelocations := []int{}

	for x, cell := range row {
		parsed, err := time.Parse("2006-1-2", cell)

		if err == nil {
			datemap[x] = parsed
			datelocations = append(datelocations, x)
		}
	}

	if len(datelocations) < 10 {
		return nil, false
	}

	for i, loc := range datelocations[:len(datelocations)-1] {
		if datemap[loc].Add(24*time.Hour) != datemap[datelocations[i+1]] {
			return nil, false
		}
	}

	return datemap, true
}
