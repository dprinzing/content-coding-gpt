package data

import (
	"encoding/csv"
	"fmt"
	"os"
)

// IdentifyCSVFile identifies a CSV file by its header.
func IdentifyCSVFile(path string) (string, error) {
	fileType := "unknown"

	// Open a CSV file reader:
	f, err := os.Open(path)
	if err != nil {
		return fileType, fmt.Errorf("identify csv file %s: %w", path, err)
	}
	defer f.Close()
	r := csv.NewReader(f)

	// Read the header:
	header, err := r.Read()
	if err != nil {
		return fileType, fmt.Errorf("identify csv file %s: %w", path, err)
	}

	// Identify the file type:
	if len(header) == 9 && header[0] == "pid" {
		fileType = "humility"
	} else if len(header) == 8 && header[0] == "pid" {
		fileType = "spiritual"
	} else if len(header) == 5 && header[0] == "pid" {
		fileType = "essay"
	}
	return fileType, nil
}

// ReadCSVFile reads a CSV file and returns the records.
func ReadCSVFile(path string, skipHeader bool) ([][]string, error) {
	// Open a CSV file reader:
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read csv file %s: %w", path, err)
	}
	defer f.Close()
	r := csv.NewReader(f)

	// Skip the header if specified:
	if skipHeader {
		_, err := r.Read()
		if err != nil {
			return nil, fmt.Errorf("read csv file %s: %w", path, err)
		}
	}

	// Read and return the records:
	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read csv file %s: %w", path, err)
	}
	return records, nil
}

// ReadCSVRecords reads a CSV file and returns a slice of records generated
// with the provided constructor. This is best-effort; errors are logged and
// broken records are ignored.
func ReadCSVRecords[T any](csvFilePath string, newRecord func([]string) (T, error)) ([]T, error) {
	var records []T

	// Read the CSV file:
	csvRecords, err := ReadCSVFile(csvFilePath, true)
	if err != nil {
		return records, err
	}

	// Convert the CSV records to records:
	for i, csvRecord := range csvRecords {
		record, err := newRecord(csvRecord)
		if err != nil {
			fmt.Printf("%s: line %d: %v\n", csvFilePath, i+2, err)
		} else {
			records = append(records, record)
		}
	}
	return records, nil
}

// WriteCSVFile writes a CSV file with the specified records.
func WriteCSVFile(path string, records [][]string) error {
	// Open a CSV file writer:
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("write csv file %s: %w", path, err)
	}
	defer f.Close()
	w := csv.NewWriter(f)

	// Write the records:
	err = w.WriteAll(records)
	if err != nil {
		return fmt.Errorf("write csv file %s: %w", path, err)
	}
	return nil
}
