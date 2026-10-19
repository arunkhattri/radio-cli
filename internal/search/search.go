// Package search provides station lookup and query capabilities using the Radio Browser API.
package search

import (
	"fmt"

	"gitlab.com/AgentNemo/goradios"
)

type StationResult struct {
	Name string
	URL  string
}

// SearchStations searches for stations by name using goradios.
func SearchStations(query string) ([]StationResult, error) {
	rawStations := goradios.FetchStations(
		goradios.StationsByName,
		query,
	)

	if len(rawStations) == 0 {
		return nil, fmt.Errorf("no stations found for query: %s", query)
	}

	var results []StationResult
	for _, st := range rawStations {
		if st.URL != "" && st.Name != "" {
			results = append(results, StationResult{
				Name: st.Name,
				URL:  st.URL,
			})
		}
	}

	return results, nil
}
