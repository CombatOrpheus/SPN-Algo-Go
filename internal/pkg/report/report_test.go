package report

import (
	"bytes"
	"spn-benchmark-ds/internal/pkg/analysis"
	"strings"
	"testing"
)

func TestCalculateStats(t *testing.T) {
	results := []*SampleResult{
		{
			NumPlaces:      2,
			NumTransitions: 1,
			Analysis: &analysis.SPNAnalysisResult{
				AverageMarkings:  []float64{0.5, 0.5},
				SteadyStateProbs: []float64{0.5, 0.5},
			},
		},
		{
			NumPlaces:      3,
			NumTransitions: 2,
			Analysis: &analysis.SPNAnalysisResult{
				AverageMarkings:  []float64{0.2, 0.8},
				SteadyStateProbs: []float64{0.2, 0.8},
			},
		},
	}

	stats := CalculateStats(results)

	if stats.NumSamples != 2 {
		t.Errorf("Expected NumSamples to be 2, but got %d", stats.NumSamples)
	}
	if stats.AvgPlaces != 2.5 {
		t.Errorf("Expected AvgPlaces to be 2.5, but got %f", stats.AvgPlaces)
	}
	if stats.AvgTransitions != 1.5 {
		t.Errorf("Expected AvgTransitions to be 1.5, but got %f", stats.AvgTransitions)
	}
	if stats.AvgMarkings != 1.0 {
		t.Errorf("Expected AvgMarkings to be 1.0, but got %f", stats.AvgMarkings)
	}
	if stats.AvgSteadyStateProbs != 1.0 {
		t.Errorf("Expected AvgSteadyStateProbs to be 1.0, but got %f", stats.AvgSteadyStateProbs)
	}
}

func TestGenerateReport(t *testing.T) {
	stats := &Stats{
		NumSamples:         10,
		AvgPlaces:          5.5,
		AvgTransitions:     3.2,
		AvgMarkings:        8.1,
		AvgSteadyStateProbs: 1.0,
	}

	var buffer bytes.Buffer
	err := GenerateReport(&buffer, stats)
	if err != nil {
		t.Fatalf("Error generating report: %v", err)
	}

	report := buffer.String()
	if !strings.Contains(report, "<td>10</td>") {
		t.Errorf("Report does not contain the correct number of samples")
	}
	if !strings.Contains(report, "<td>5.5</td>") {
		t.Errorf("Report does not contain the correct average number of places")
	}
}
