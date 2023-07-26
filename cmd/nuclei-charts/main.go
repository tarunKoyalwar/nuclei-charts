package main

import (
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/projectdiscovery/gologger"
	nucleicharts "github.com/tarunKoyalwar/nuclei-charts"
)

var output = ""
var input = ""

func main() {

	flag.StringVar(&output, "output", "", "Output HTML file to save Page (includes all charts)")
	flag.StringVar(&input, "input", "stats.json", "Input JSON file to read stats from")
	flag.Parse()

	stats, err := nucleicharts.ReadStatsFromFile(input)
	if err != nil {
		panic(err)
	}

	if output != "" {
		page := nucleicharts.AllCharts(stats, time.Second)
		f, err := os.Create(output)
		if err != nil {
			panic(err)
		}
		page.Render(f)
		f.Close()
		return
	}

	http.HandleFunc("/line", func(w http.ResponseWriter, r *http.Request) {
		line := nucleicharts.TotalRequestsOverTime(stats)
		line.Render(w)
	})
	http.HandleFunc("/kline", func(w http.ResponseWriter, r *http.Request) {
		line := nucleicharts.TopSlowTemplates(stats)
		line.Render(w)
	})
	http.HandleFunc("/rps", func(w http.ResponseWriter, r *http.Request) {
		line := nucleicharts.RequestsVSInterval(stats, getTimeWithFallback(r.URL.Query().Get("interval")))
		line.Render(w)
	})

	http.HandleFunc("/concurrency", func(w http.ResponseWriter, r *http.Request) {
		line := nucleicharts.ConcurrencyVsTime(stats, getTimeWithFallback(r.URL.Query().Get("interval")))
		line.Render(w)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		page := nucleicharts.AllCharts(stats, getTimeWithFallback(r.URL.Query().Get("interval")))
		page.Render(w)
	})

	gologger.Info().Msgf("Starting server on http://localhost:8081/?interval=1s")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		gologger.Error().Msgf("%v", err)
	}
}

func getTimeWithFallback(interval string) time.Duration {
	d, err := time.ParseDuration(interval)
	if err != nil {
		gologger.Error().Msgf("Could not parse interval %s: %s\n", interval, err)
		d = time.Second
	}
	return d
}
