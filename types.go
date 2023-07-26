package nucleicharts

import (
	"encoding/json"
	"os"
	"time"
)

type ItemType string

const (
	ItemStart ItemType = "start"
	ItemEnd   ItemType = "end"
)

type Item struct {
	ID           string
	Time         time.Time
	TemplateType string
	Target       string
	ItemType     ItemType
	Requests     int
}

// Stats is the stats object
type Stats struct {
	TemplateStart []Item `json:"template-start"`
	TemplateEnd   []Item `json:"template-end"`
	Concurrency   int    `json:"concurrency"`
}

// Save saves the stats to a file
func (s Stats) Save() error {
	filename := "stats.json"
	if val := os.Getenv("NUCLEI_STATS_FILE"); val != "" {
		filename = val
	}
	bin, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, bin, 0644)
}

// ReadStatsFromFile reads stats from a file
func ReadStatsFromFile(filename string) (*Stats, error) {
	bin, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	stats := &Stats{}
	err = json.Unmarshal(bin, stats)
	if err != nil {
		return nil, err
	}
	return stats, nil
}
