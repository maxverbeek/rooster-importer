package domain

import (
	"fmt"
	"strings"
	"time"
)

type ScheduleEvent struct {
	ScheduleType string
	Start        time.Time
	End          time.Time
	AllDay       bool
}

func (e *ScheduleEvent) Summary() string {
	return fmt.Sprintf("%s: %s (%s - %s)", e.ScheduleType, e.Start.Format("02/01"), e.Start.Format("15:04"), e.End.Format("15:04"))
}

type Conversion string

const (
	ConversionVrij      Conversion = "vrij"
	ConversionConverted Conversion = "converted"
	ConversionDefaulted Conversion = "defaulted"
	ConversionSkipped   Conversion = "skipped"
)

func timeAtDay(date time.Time, hours, minutes int) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), hours, minutes, 0, 0, time.Local)
}

func dateToTime(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
}

func NewScheduleEvent(excelEntry string, date time.Time) (*ScheduleEvent, Conversion) {
	weekday := timeAtDay(date, 16, 0).Weekday()
	isWeekend := weekday == time.Saturday || weekday == time.Sunday

	switch strings.ToLower(strings.Trim(excelEntry, " ")) {
	case "":
		fallthrough
	case "x":
		fallthrough
	case "-":
		fallthrough
	case "-c":
		conversion := ConversionVrij
		if isWeekend {
			conversion = ConversionSkipped
		}

		event := ScheduleEvent{
			ScheduleType: "Vrij",
			Start:        dateToTime(date),
			End:          dateToTime(date.Add(24 * time.Hour)),
			AllDay:       true,
		}
		return &event, conversion

	case "d":
		fallthrough
	case "d (als)":
		event := ScheduleEvent{
			ScheduleType: "Dag",
			Start:        timeAtDay(date, 7, 45),
			End:          timeAtDay(date, 16, 15),
			AllDay:       false,
		}
		return &event, ConversionConverted

	case "t":
		fallthrough
	case "t (als)":
		event := ScheduleEvent{
			ScheduleType: "Tussen",
			Start:        timeAtDay(date, 11, 0),
			End:          timeAtDay(date, 19, 30),
			AllDay:       false,
		}
		return &event, ConversionConverted

	case "a":
		event := ScheduleEvent{
			ScheduleType: "Avond",
			Start:        timeAtDay(date, 15, 0),
			End:          timeAtDay(date, 23, 30),
			AllDay:       false,
		}
		return &event, ConversionConverted

	case "n":
		event := ScheduleEvent{
			ScheduleType: "Nacht",
			Start:        timeAtDay(date, 23, 0),
			End:          timeAtDay(date.Add(24*time.Hour), 8, 30),
			AllDay:       false,
		}
		return &event, ConversionConverted

	case "vak":
		fallthrough
	case "vak.":
		event := ScheduleEvent{
			ScheduleType: "Vakantie",
			Start:        dateToTime(date),
			End:          dateToTime(date.Add(24 * time.Hour)),
			AllDay:       true,
		}
		return &event, ConversionConverted

	}

	// default naar dagdienst met een waarschuwing als het roostertype niet herkent wordt.
	return &ScheduleEvent{
		ScheduleType: "Dag",
		Start:        timeAtDay(date, 7, 45),
		End:          timeAtDay(date, 16, 15),
		AllDay:       false,
	}, ConversionDefaulted

	// case "":
	// 	// weekend dagen niet meenemen
	// 	return nil, ConversionSkipped
	// case "aanvraag verlof":
	// 	fallthrough
	// case "aanvraag vrij":
	// 	fallthrough
	// case "vak":
	// 	fallthrough
	// case "vk":
	// 	// Vakantie dag
	// 	event := ScheduleEvent{
	// 		ScheduleType: "Vakantiedag",
	// 		Start:        dateToTime(date),
	// 		End:          dateToTime(date.Add(24 * time.Hour)),
	// 		AllDay:       true,
	// 	}
	//
	// 	return &event, ConversionVrij
	// case "c":
	// 	// Vrij/Compensatie
	// 	event := ScheduleEvent{
	// 		ScheduleType: "Compensatiedag",
	// 		Start:        dateToTime(date),
	// 		End:          dateToTime(date.Add(24 * time.Hour)),
	// 		AllDay:       true,
	// 	}
	//
	// 	return &event, ConversionVrij
	//
	// case "a":
	// 	fallthrough
	// case "wa":
	// 	// (weekend) avond dienst
	// 	event := ScheduleEvent{
	// 		ScheduleType: "Avond",
	// 		Start:        timeAtDay(date, 16, 0),
	// 		End:          timeAtDay(date, 23, 59),
	// 	}
	//
	// 	return &event, ConversionConverted
	//
	// case "n":
	// 	fallthrough
	// case "wn":
	// 	// (weekend) nachtdienst
	// 	event := ScheduleEvent{
	// 		ScheduleType: "Nacht",
	// 		Start:        timeAtDay(date, 23, 30),
	// 		End:          timeAtDay(date.Add(time.Hour*24), 8, 30),
	// 	}
	//
	// 	return &event, ConversionConverted
	//
	// case "wk":
	// 	// weekend kort
	// 	event := ScheduleEvent{
	// 		ScheduleType: "Weekend kort",
	// 		Start:        timeAtDay(date, 8, 0),
	// 		End:          timeAtDay(date, 13, 0),
	// 	}
	//
	// 	return &event, ConversionConverted
	//
	// case "1b":
	// 	fallthrough
	// case "1d":
	// 	fallthrough
	// case "ccu":
	// 	fallthrough
	// case "seh":
	// 	fallthrough
	// case "seh l":
	// 	fallthrough
	// case "seh c":
	// 	// Normale dienst
	// 	event := ScheduleEvent{
	// 		ScheduleType: excelEntry,
	// 		Start:        timeAtDay(date, 8, 0),
	// 		End:          timeAtDay(date, 17, 0),
	// 	}
	//
	// 	return &event, ConversionConverted
	// }
	//
	// // Fall back to 8 - 17 but with a warning
	// return &ScheduleEvent{
	// 	ScheduleType: excelEntry,
	// 	Start:        timeAtDay(date, 8, 0),
	// 	End:          timeAtDay(date, 17, 0),
	// }, ConversionDefaulted
}
