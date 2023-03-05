package data

import (
	"content-coding-gpt/pkg/openai"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var HumilityCSVHeader = []string{"pid", "response", "hum1", "hum2", "hum3", "hum4", "hum5", "hum6", "standardized"}

// HumilityRecord is a training record from the humility csv files.
// pid,response,hum1,hum2,hum3,hum4,hum5,hum6,standardized
type HumilityRecord struct {
	ID       int     `csv:"pid"`
	Response string  `csv:"response"`
	S1       int     `csv:"hum1"`
	S2       int     `csv:"hum2"`
	S3       int     `csv:"hum3"`
	S4       int     `csv:"hum4"`
	S5       int     `csv:"hum5"`
	S6       int     `csv:"hum6"`
	Std      float64 `csv:"standardized"`
}

// CSVHeader returns the header for a HumilityRecord as a slices of strings.
func (r HumilityRecord) CSVHeader() []string {
	return HumilityCSVHeader
}

// CSVFields returns the fields for a HumilityRecord as a slices of strings.
func (r HumilityRecord) CSVFields() []string {
	return []string{
		strconv.Itoa(r.ID),
		r.Response,
		strconv.Itoa(r.S1),
		strconv.Itoa(r.S2),
		strconv.Itoa(r.S3),
		strconv.Itoa(r.S4),
		strconv.Itoa(r.S5),
		strconv.Itoa(r.S6),
		fmt.Sprintf("%.2f", r.Std),
	}
}

// Results returns the results of the HumilityRecord.
func (r HumilityRecord) Results() string {
	return fmt.Sprintf("%d %d %d %d %d %d %.2f", r.S1, r.S2, r.S3, r.S4, r.S5, r.S6, r.Std)
}

// PlainTrainingRecord converts a HumilityRecord to a plain TrainingRecord.
func (r HumilityRecord) PlainTrainingRecord() TrainingRecord {
	return TrainingRecord{
		Prompt: r.Response + PromptSeparator,
		Completion: CompletionStart +
			fmt.Sprintf("%d %d %d %d %d %d %.2f", r.S1, r.S2, r.S3, r.S4, r.S5, r.S6, r.Std) +
			CompletionStop,
	}
}

// NewHumilityRecordCSV creates a new HumilityRecord from a slice of strings,
// ostensibly read from a CSV file.
func NewHumilityRecordCSV(fields []string) (HumilityRecord, error) {
	var record HumilityRecord
	var err error
	if len(fields) != 9 {
		return record, errors.New("invalid number of fields")
	}
	record.ID, err = strconv.Atoi(fields[0])
	if err != nil {
		return record, fmt.Errorf("invalid pid %s: %w", fields[0], err)
	}
	record.Response = CleanResponse(fields[1])
	if record.Response == "" {
		return record, errors.New("empty response")
	}
	record.S1, err = strconv.Atoi(fields[2])
	if err != nil {
		return record, fmt.Errorf("invalid hum1 %s: %w", fields[2], err)
	}
	record.S2, err = strconv.Atoi(fields[3])
	if err != nil {
		return record, fmt.Errorf("invalid hum2 %s: %w", fields[3], err)
	}
	record.S3, err = strconv.Atoi(fields[4])
	if err != nil {
		return record, fmt.Errorf("invalid hum3 %s: %w", fields[4], err)
	}
	record.S4, err = strconv.Atoi(fields[5])
	if err != nil {
		return record, fmt.Errorf("invalid hum4 %s: %w", fields[5], err)
	}
	record.S5, err = strconv.Atoi(fields[6])
	if err != nil {
		return record, fmt.Errorf("invalid hum5 %s: %w", fields[6], err)
	}
	record.S6, err = strconv.Atoi(fields[7])
	if err != nil {
		return record, fmt.Errorf("invalid hum6 %s: %w", fields[7], err)
	}
	record.Std, err = strconv.ParseFloat(fields[8], 64)
	if err != nil {
		return record, fmt.Errorf("invalid standardized %s: %w", fields[8], err)
	}
	return record, nil
}

// NewHumilityRecord creates a new HumilityRecord from an EssayRecord and a Completion.
func NewHumilityRecord(e EssayRecord, essayType string, c openai.Completion) (HumilityRecord, error) {
	var r HumilityRecord
	r.ID = e.ID
	r.Response = e.SelectEssay(essayType)
	if len(c.Choices) >= 0 {
		fields := strings.Fields(c.Choices[0].Text)
		if len(fields) > 0 {
			s1, err := strconv.Atoi(fields[0])
			if err != nil {
				return r, fmt.Errorf("invalid hum1 %s: %w", fields[0], err)
			}
			r.S1 = s1
		}
		if len(fields) > 1 {
			s2, err := strconv.Atoi(fields[1])
			if err != nil {
				return r, fmt.Errorf("invalid hum2 %s: %w", fields[1], err)
			}
			r.S2 = s2
		}
		if len(fields) > 2 {
			s3, err := strconv.Atoi(fields[2])
			if err != nil {
				return r, fmt.Errorf("invalid hum3 %s: %w", fields[2], err)
			}
			r.S3 = s3
		}
		if len(fields) > 3 {
			s4, err := strconv.Atoi(fields[3])
			if err != nil {
				return r, fmt.Errorf("invalid hum4 %s: %w", fields[3], err)
			}
			r.S4 = s4
		}
		if len(fields) > 4 {
			s5, err := strconv.Atoi(fields[4])
			if err != nil {
				return r, fmt.Errorf("invalid hum5 %s: %w", fields[4], err)
			}
			r.S5 = s5
		}
		if len(fields) > 5 {
			s6, err := strconv.Atoi(fields[5])
			if err != nil {
				return r, fmt.Errorf("invalid hum6 %s: %w", fields[5], err)
			}
			r.S6 = s6
		}
		if len(fields) > 6 {
			std, err := strconv.ParseFloat(fields[6], 64)
			if err != nil {
				return r, fmt.Errorf("invalid standardized %s: %w", fields[6], err)
			}
			r.Std = std
		}
	}
	return r, nil
}

// ReadHumilityRecords reads a CSV file and returns a slice of HumilityRecords.
// This is best-effort; errors are logged and broken records are ignored.
func ReadHumilityRecords(path string) ([]HumilityRecord, error) {
	return ReadCSVRecords(path, NewHumilityRecordCSV)
}

// WriteHumilityRecords writes a slice of HumilityRecords to a CSV file.
func WriteHumilityRecords(path string, records []HumilityRecord) error {
	csvRecords := make([][]string, len(records)+1)
	csvRecords[0] = HumilityCSVHeader
	for i, r := range records {
		csvRecords[i+1] = r.CSVFields()
	}
	return WriteCSVFile(path, csvRecords)
}
