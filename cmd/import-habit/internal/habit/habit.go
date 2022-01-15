package habit

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"time"
)

const (
	fieldHeaderHabit = "Habit"
	fieldHeaderDate  = "Date"
)

func ReadEntriesFromFile(filePath string) (map[string][]time.Time, error) {
	csvContents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Habit CSV file: %w", err)
	}

	result := make(map[string][]time.Time)
	r := csv.NewReader(bytes.NewReader(csvContents))
	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read header CSV record: %w", err)
	}
	habitIndex := -1
	dateIndex := -1
	for i, heading := range header {
		if heading == fieldHeaderHabit {
			habitIndex = i
		}
		if heading == fieldHeaderDate {
			dateIndex = i
		}
	}
	if habitIndex == -1 || dateIndex == -1 {
		return nil, fmt.Errorf("failed to find required columns Habit and Date in CSV")
	}
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to parse record: %w", err)
		}
		habit := record[habitIndex]
		date, err := time.Parse("2006-01-02", record[dateIndex])
		if err != nil {
			return nil, fmt.Errorf("failed to parse date: %w", err)
		}
		if times, ok := result[habit]; ok {
			result[habit] = append(times, date)
		} else {
			result[habit] = []time.Time{date}
		}
	}
	return result, nil
}
