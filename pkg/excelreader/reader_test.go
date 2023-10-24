package excelreader_test

import (
	"fmt"
	"os"
	"rooster-importer/pkg/excelreader"
	"testing"
	"time"
)

func TestHandleSelectedFile(t *testing.T) {
	file, err := os.Open("/home/max/Downloads/Rooster ANIOS cardio-long 2023 - KOPIE (1).xlsx")

	if err != nil {
		t.Log("wtf?")
		t.Error(err)
	}

	entries, err := excelreader.HandleSelectedFile(file, "Nerea")

	for _, entry := range entries {
		fmt.Printf("%s: %s\n", entry.Date.Format(time.DateOnly), entry.Shift)
	}

	if err != nil {
		t.Error(err)
	}
}

func TestDateParsing(t *testing.T) {
	parsed, err := time.Parse("2006-1-2", "2023-1-17")

	if err != nil {
		t.Fatal(err)
	}

	year, month, day := parsed.Date()

	if day != 17 {
		t.Errorf("wrong day (expected 17, is %d): year: %d, month: %s", day, year, month)
	}
}
