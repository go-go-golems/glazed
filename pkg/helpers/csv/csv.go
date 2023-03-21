package csv

import (
	"encoding/csv"
	"io"
	"strconv"
)

type ParseCSVOption func(*csv.Reader)

func WithComma(c rune) ParseCSVOption {
	return func(r *csv.Reader) {
		r.Comma = c
	}
}

func WithComment(c rune) ParseCSVOption {
	return func(r *csv.Reader) {
		r.Comment = c
	}
}

func WithLazyQuotes(l bool) ParseCSVOption {
	return func(r *csv.Reader) {
		r.LazyQuotes = l
	}
}

func WithTrimLeadingSpace(t bool) ParseCSVOption {
	return func(r *csv.Reader) {
		r.TrimLeadingSpace = t
	}
}

func WithFieldsPerRecord(f int) ParseCSVOption {
	return func(r *csv.Reader) {
		r.FieldsPerRecord = f
	}
}

func ParseCSV(r io.Reader, options ...ParseCSVOption) (
	[]map[string]interface{}, error,
) {
	csvReader := csv.NewReader(r)

	for _, o := range options {
		o(csvReader)
	}

	// Read the header row of the CSV file to use as keys for the maps
	header, err := csvReader.Read()
	if err != nil {
		return nil, err
	}

	var data []map[string]interface{}

	for {
		// Read each row of the CSV file
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// Create a new map to store the row data
		rowData := make(map[string]interface{})

		// Iterate over each column in the row and store the value in the map
		for i, value := range row {
			var v interface{}
			v = value
			// check if we can cast to int
			if intV, err := strconv.Atoi(value); err == nil {
				v = intV
			} else if floatV, err := strconv.ParseFloat(value, 64); err == nil {
				v = floatV
			}

			rowData[header[i]] = v
		}

		// Add the row data to the slice of maps
		data = append(data, rowData)
	}

	return data, nil
}
