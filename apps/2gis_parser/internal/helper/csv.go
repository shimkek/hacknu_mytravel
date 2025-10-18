package helper

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

func LoadKeywordsFromCSV(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open keywords file %s: %w", filename, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read keywords CSV: %w", err)
	}

	var keywords []string
	for _, record := range records {
		if len(record) > 0 && strings.TrimSpace(record[0]) != "" {
			keywords = append(keywords, strings.TrimSpace(record[0]))
		}
	}
	return keywords, nil
}

// LoadFieldsFromCSV loads field names from a CSV file
func LoadFieldsFromCSV(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open fields file %s: %w", filename, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read fields CSV: %w", err)
	}

	var fields []string
	for _, record := range records {
		if len(record) > 0 && strings.TrimSpace(record[0]) != "" {
			fields = append(fields, strings.TrimSpace(record[0]))
		}
	}
	return fields, nil
}
