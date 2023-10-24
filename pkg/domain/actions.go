package domain

import (
	"io"
	"rooster-importer/pkg/excelreader"
)

type Action func(*Application)

func SelectedXlsxFileAction(file io.ReadCloser) Action {
	return func(a *Application) {
		a.xlsxfile = file
		_, err := excelreader.HandleSelectedFile(file, "Nerea")

		if err != nil {
			a.guistuff <- err
		}
	}
}
