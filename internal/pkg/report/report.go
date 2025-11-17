package report

import (
	"html/template"
	"io"
	"spn-benchmark-ds/internal/pkg/analysis"
)

// SampleResult holds the analysis result for a single sample.
type SampleResult struct {
	NumPlaces        int
	NumTransitions   int
	Analysis         *analysis.SPNAnalysisResult
}

// Stats holds the statistics for the generated dataset.
type Stats struct {
	NumSamples         int
	AvgPlaces          float64
	AvgTransitions     float64
	AvgMarkings        float64
	AvgSteadyStateProbs float64
}

// GenerateReport generates an HTML report from the given stats.
func GenerateReport(writer io.Writer, stats *Stats) error {
	tmpl, err := template.New("report").Parse(reportTemplate)
	if err != nil {
		return err
	}
	return tmpl.Execute(writer, stats)
}

// CalculateStats calculates the statistics for the generated dataset.
func CalculateStats(results []*SampleResult) *Stats {
	if len(results) == 0 {
		return &Stats{}
	}

	stats := &Stats{
		NumSamples: len(results),
	}

	var totalPlaces, totalTransitions, totalMarkings, totalSteadyStateProbs float64
	for _, result := range results {
		totalPlaces += float64(result.NumPlaces)
		totalTransitions += float64(result.NumTransitions)
		totalMarkings += sumFloat64(result.Analysis.AverageMarkings)
		totalSteadyStateProbs += sumFloat64(result.Analysis.SteadyStateProbs)
	}

	stats.AvgPlaces = totalPlaces / float64(len(results))
	stats.AvgTransitions = totalTransitions / float64(len(results))
	stats.AvgMarkings = totalMarkings / float64(len(results))
	stats.AvgSteadyStateProbs = totalSteadyStateProbs / float64(len(results))

	return stats
}

func sumFloat64(slice []float64) float64 {
	var sum float64
	for _, v := range slice {
		sum += v
	}
	return sum
}

const reportTemplate = `
<!DOCTYPE html>
<html>
<head>
	<title>SPN Dataset Statistics</title>
	<style>
		body { font-family: sans-serif; }
		table { border-collapse: collapse; }
		th, td { border: 1px solid #ddd; padding: 8px; }
		th { background-color: #f2f2f2; }
	</style>
</head>
<body>
	<h1>SPN Dataset Statistics</h1>
	<table>
		<tr>
			<th>Statistic</th>
			<th>Value</th>
		</tr>
		<tr>
			<td>Number of samples</td>
			<td>{{.NumSamples}}</td>
		</tr>
		<tr>
			<td>Average number of places</td>
			<td>{{.AvgPlaces}}</td>
		</tr>
		<tr>
			<td>Average number of transitions</td>
			<td>{{.AvgTransitions}}</td>
		</tr>
		<tr>
			<td>Average number of markings</td>
			<td>{{.AvgMarkings}}</td>
		</tr>
		<tr>
			<td>Average sum of steady state probabilities</td>
			<td>{{.AvgSteadyStateProbs}}</td>
		</tr>
	</table>
</body>
</html>
`
