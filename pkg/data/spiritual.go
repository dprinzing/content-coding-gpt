package data

import (
	"content-coding-gpt/pkg/openai"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var SpiritualCSVHeader = []string{"pid", "response", "layDefinition", "spir1", "spir2", "spir3", "spir4", "standardized"}

// SpiritualRecord is a training record from the spiritual csv files.
// pid,response,layDefinition,spir1,spir2,spir3,spir4,standardized
type SpiritualRecord struct {
	ID            int     `csv:"pid"`
	Response      string  `csv:"response"`
	LayDefinition int     `csv:"layDefinition"`
	S1            int     `csv:"spir1"`
	S2            int     `csv:"spir2"`
	S3            int     `csv:"spir3"`
	S4            int     `csv:"spir4"`
	Std           float64 `csv:"standardized"`
}

// CSVHeader returns the header for a SpiritualRecord as a slices of strings.
func (r SpiritualRecord) CSVHeader() []string {
	return SpiritualCSVHeader
}

// CSVFields returns the fields for a SpiritualRecord as a slices of strings.
func (r SpiritualRecord) CSVFields() []string {
	return []string{
		strconv.Itoa(r.ID),
		r.Response,
		strconv.Itoa(r.LayDefinition),
		strconv.Itoa(r.S1),
		strconv.Itoa(r.S2),
		strconv.Itoa(r.S3),
		strconv.Itoa(r.S4),
		fmt.Sprintf("%.2f", r.Std),
	}
}

// Results returns the results of the SpiritualRecord.
func (r SpiritualRecord) Results() string {
	return fmt.Sprintf("%d %d %d %d %d %.2f", r.LayDefinition, r.S1, r.S2, r.S3, r.S4, r.Std)
}

// PlainTrainingRecord converts a SpiritualRecord to a plain TrainingRecord.
func (r SpiritualRecord) PlainTrainingRecord() TrainingRecord {
	return TrainingRecord{
		Prompt: r.Response + PromptSeparator,
		Completion: CompletionStart +
			fmt.Sprintf("%d %d %d %d %d %.2f", r.LayDefinition, r.S1, r.S2, r.S3, r.S4, r.Std) +
			CompletionStop,
	}
}

// NewSpiritualRecordCSV creates a new SpiritualRecord from a slice of strings,
// ostensibly read from a CSV file.
func NewSpiritualRecordCSV(fields []string) (SpiritualRecord, error) {
	var record SpiritualRecord
	var err error
	if len(fields) != 8 {
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
	record.LayDefinition, err = strconv.Atoi(fields[2])
	if err != nil {
		return record, fmt.Errorf("invalid layDefinition %s: %w", fields[2], err)
	}
	record.S1, err = strconv.Atoi(fields[3])
	if err != nil {
		return record, fmt.Errorf("invalid spir1 %s: %w", fields[3], err)
	}
	record.S2, err = strconv.Atoi(fields[4])
	if err != nil {
		return record, fmt.Errorf("invalid spir2 %s: %w", fields[4], err)
	}
	record.S3, err = strconv.Atoi(fields[5])
	if err != nil {
		return record, fmt.Errorf("invalid spir3 %s: %w", fields[5], err)
	}
	record.S4, err = strconv.Atoi(fields[6])
	if err != nil {
		return record, fmt.Errorf("invalid spir4 %s: %w", fields[6], err)
	}
	record.Std, err = strconv.ParseFloat(fields[7], 64)
	if err != nil {
		return record, fmt.Errorf("invalid standardized %s: %w", fields[7], err)
	}
	return record, nil
}

// NewSpiritualRecord creates a new SpiritualRecord from an EssayRecord and a Completion.
func NewSpiritualRecord(e EssayRecord, essayType string, c openai.Completion) (SpiritualRecord, error) {
	var r SpiritualRecord
	r.ID = e.ID
	r.Response = e.SelectEssay(essayType)
	if len(c.Choices) >= 0 {
		fields := strings.Fields(c.Choices[0].Text)
		if len(fields) > 0 {
			layDefinition, err := strconv.Atoi(fields[0])
			if err != nil {
				return r, fmt.Errorf("invalid layDefinition %s: %w", fields[0], err)
			}
			r.LayDefinition = layDefinition
		}
		if len(fields) > 1 {
			s1, err := strconv.Atoi(fields[1])
			if err != nil {
				return r, fmt.Errorf("invalid spir1 %s: %w", fields[1], err)
			}
			r.S1 = s1
		}
		if len(fields) > 2 {
			s2, err := strconv.Atoi(fields[2])
			if err != nil {
				return r, fmt.Errorf("invalid spir2 %s: %w", fields[2], err)
			}
			r.S2 = s2
		}
		if len(fields) > 3 {
			s3, err := strconv.Atoi(fields[3])
			if err != nil {
				return r, fmt.Errorf("invalid spir3 %s: %w", fields[3], err)
			}
			r.S3 = s3
		}
		if len(fields) > 4 {
			s4, err := strconv.Atoi(fields[4])
			if err != nil {
				return r, fmt.Errorf("invalid spir4 %s: %w", fields[4], err)
			}
			r.S4 = s4
		}
		if len(fields) > 5 {
			std, err := strconv.ParseFloat(fields[5], 64)
			if err != nil {
				return r, fmt.Errorf("invalid standardized %s: %w", fields[5], err)
			}
			r.Std = std
		}
	}
	return r, nil
}

// ReadSpiritualRecords reads a CSV file and returns a slice of SpiritualRecords.
// This is best-effort; errors are logged and broken records are ignored.
func ReadSpiritualRecords(path string) ([]SpiritualRecord, error) {
	return ReadCSVRecords(path, NewSpiritualRecordCSV)
}

// WriteSpiritualRecords writes a slice of SpiritualRecords to a CSV file.
func WriteSpiritualRecords(path string, records []SpiritualRecord) error {
	csvRecords := make([][]string, len(records)+1)
	csvRecords[0] = SpiritualCSVHeader
	for i, r := range records {
		csvRecords[i+1] = r.CSVFields()
	}
	return WriteCSVFile(path, csvRecords)
}
