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

type NoEntriesFoundError struct {
	sheets []string
}

func (e *NoEntriesFoundError) Error() string {

	if len(e.sheets) == 1 {
		return fmt.Sprintf("didn't find entries in sheet %s", e.sheets[0])
	}

	str := strings.Builder{}

	str.WriteString("didn't find entries in sheets ")

	for i, sheet := range e.sheets {
		str.WriteString(sheet)

		if i < len(e.sheets)-1 {
			str.WriteString(", ")
		}
	}

	return str.String()
}

var DateParseError error = errors.New("couldn't parse date")
var NoEntriesInSheet error = errors.New("no entries found")
var NotAScheduleSheet error = errors.New("sheet does not contain dates")

func FindScheduleEntries(reader io.ReadCloser, name string) ([]ScheduleEntry, error) {
	file, err := excelize.OpenReader(reader)

	if err != nil {
		return nil, err
	}

	sheets := file.GetSheetList()

	if len(sheets) == 0 {
		return nil, errors.New("No sheets found in the Excel file")
	}

	allEntries := []ScheduleEntry{}

	noEntriesError := &NoEntriesFoundError{}

	for _, sheet := range sheets {
		entries, err := processSheet(file, sheet, name)

		if err != nil {
			if errors.Is(err, NoEntriesInSheet) || errors.Is(err, NotAScheduleSheet) {
				noEntriesError.sheets = append(noEntriesError.sheets, sheet)
			} else {
				return nil, fmt.Errorf("error in processing sheet %s: %w", sheet, err)
			}
		}

		allEntries = append(allEntries, entries...)
	}

	// ultimately, return all of the entries + optionally an error if any sheets did not yield any entries.
	// if there are no sheets where we have a no entry found error, don't return an error at all
	if len(noEntriesError.sheets) == 0 {
		return allEntries, nil
	}

	return allEntries, noEntriesError
}

func processSheet(file *excelize.File, sheet string, name string) ([]ScheduleEntry, error) {
	rows, err := file.GetRows(sheet)

	if err != nil {
		return nil, fmt.Errorf("error in iterating over rows: %w", err)
	}

	var datemapping map[int]time.Time

	for rowidx, row := range rows {
		if len(row) == 0 {
			continue
		}

		// While iterating over rows, check if the current row contains dates
		if mapping, ok := findDateRow(row, rowidx, file, sheet); ok {
			datemapping = mapping
		}

		if strings.HasPrefix(row[0], name) {
			if datemapping == nil {
				return nil, fmt.Errorf("found %s before knowing the dates: %w", name, NotAScheduleSheet)
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

	// if we get here, the loop above didn't yield any entries..
	// return a "no entries found" error
	return nil, NoEntriesInSheet
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
