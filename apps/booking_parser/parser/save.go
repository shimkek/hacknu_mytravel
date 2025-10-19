package parser

import (
	"encoding/json"
	"fmt"
	"os"
)

// SaveFinalData writes the final array of parsed property details to final_data.json
func SaveFinalData(data []DetailedProperty) error {
	if len(data) == 0 {
		fmt.Println("⚠️  No data to save (final_data.json will be empty).")
	}

	file, err := os.Create("final_data.json")
	if err != nil {
		return fmt.Errorf("failed to create final_data.json: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to write JSON: %v", err)
	}

	fmt.Println("💾 Saved successfully → final_data.json")
	return nil
}
