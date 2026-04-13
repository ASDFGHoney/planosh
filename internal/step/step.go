package step

import (
	"encoding/json"
	"fmt"
	"os"
)

// Verify represents a single verification command for a step.
type Verify struct {
	Name string `json:"name"`
	Run  string `json:"run"`
}

// Step represents one step in a plan.
type Step struct {
	ID     int      `json:"id"`
	Name   string   `json:"name"`
	Prompt string   `json:"prompt"`
	Verify []Verify `json:"verify"`
	Commit string   `json:"commit"`
}

// Plan represents the top-level structure of steps.json.
type Plan struct {
	PlanName string `json:"plan_name"`
	PRD      string `json:"prd"`
	Created  string `json:"created"`
	Steps    []Step `json:"steps"`
}

// Parse reads a steps.json file and returns the parsed Plan.
func Parse(path string) (*Plan, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read steps.json: %w", err)
	}

	var plan Plan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, fmt.Errorf("parse steps.json: %w", err)
	}

	return &plan, nil
}
