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

var DateParseError error = errors.New("couldn't parse date")

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

	for rowidx, row := range rows {
		if len(row) == 0 {
			continue
		}

		if mapping, ok := findDateRow(row, rowidx, file, sheets[0]); ok {
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

func findDateRow(row []string, rowidx int, file *excelize.File, sheetName string) (map[int]time.Time, bool) {
	datemap := make(map[int]time.Time)
	datelocations := []int{}

	for x, cell := range row {

		// first try interpreting the date format from the stylesheet of the excel file
		parsed, err := parseUsingStylesheet(x, rowidx, file, sheetName, cell)

		if err != nil {
			// try yyyy-mm-dd
			parsed, err = time.Parse("2006-1-2", cell)
		}

		if err != nil {
			// try day-month-year format as well
			parsed, err = time.Parse("2-1-2006", cell)
		}

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

func parseUsingStylesheet(col, row int, file *excelize.File, sheetName, content string) (time.Time, error) {
	cellname, err := excelize.CoordinatesToCellName(col+1, row+1, false)

	if err != nil {
		return time.Time{}, err
	}

	style, err := file.GetCellStyle(sheetName, cellname)

	if err != nil {
		return time.Time{}, err
	}

	if stylesheet := file.Styles.CellXfs.Xf; style < len(stylesheet) {
		var format string

		if stylesheet[style].NumFmtID == nil {
			return time.Time{}, DateParseError
		}

		switch *stylesheet[style].NumFmtID {
		case 14:
			//"mm-dd-yy"
			format = "01-02-06"
		case 15:
			//"d-mmm-yy"
			format = "02-Jan-06"
		case 16:
			//"d-mmm"
			format = "02-Jan"
		case 17:
			//"mmm-yy"
			format = "Jan-06"
		case 22:
			//"m/d/yy h:mm"
			format = "1/2/06 15:04"
		default:
			return time.Time{}, DateParseError
		}

		t, err := time.Parse(format, content)

		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, DateParseError
}
