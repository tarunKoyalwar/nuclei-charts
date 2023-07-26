package nucleicharts

import (
	"fmt"
	"sort"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
)

const (
	TopK         = 50
	SpacerHeight = "50px"
)

func AllCharts(stats *Stats, duration time.Duration) *components.Page {
	page := components.NewPage()
	page.PageTitle = "Nuclei Charts"
	line1 := TotalRequestsOverTime(stats)
	line1.SetSpacerHeight(SpacerHeight)
	kline := TopSlowTemplates(stats)
	kline.SetSpacerHeight(SpacerHeight)
	line2 := RequestsVSInterval(stats, duration)
	line2.SetSpacerHeight(SpacerHeight)
	line3 := ConcurrencyVsTime(stats, duration)
	line3.SetSpacerHeight(SpacerHeight)
	page.AddCharts(line1, kline, line2, line3)
	page.Validate()
	page.SetLayout(components.PageCenterLayout)
	return page
}

func TotalRequestsOverTime(stats *Stats) *charts.Line {
	line := charts.NewLine()
	line.SetCaption("Chart Shows Total Requests Count Over Time (for each/all Protocols)")

	var startTime time.Time = time.Now()
	var endTime time.Time

	// in this plot we only want end times
	for _, item := range stats.TemplateEnd {
		if item.Time.Before(startTime) {
			startTime = item.Time
		}
		if item.Time.After(endTime) {
			endTime = item.Time
		}
	}

	data := getCategoryRequestCount(stats.TemplateEnd)
	max := 0
	for _, v := range data {
		if len(v) > max {
			max = len(v)
		}
	}
	line = line.SetXAxis(time.Now().Format(time.RFC3339))
	for k, v := range data {
		lineData := []opts.LineData{}
		temp := 0
		for _, item := range v {
			temp += item.Requests
			val := item.Time.Sub(startTime)
			lineData = append(lineData, opts.LineData{
				Value: []any{val.Milliseconds(), temp},
				Name:  item.ID,
			})
		}
		line = line.AddSeries(k, lineData, charts.WithLineChartOpts(opts.LineChart{Smooth: false}), charts.WithLabelOpts(opts.Label{Show: true, Position: "top"}))
	}

	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Nuclei: total-req vs time"}),
		charts.WithXAxisOpts(opts.XAxis{Name: "Time", Type: "time", AxisLabel: &opts.AxisLabel{Show: true, ShowMaxLabel: true, Formatter: opts.FuncOpts(`function (date) { return (date/1000)+'s'; }`)}}),
		charts.WithYAxisOpts(opts.YAxis{Name: "Requests Sent", Type: "value"}),
		charts.WithInitializationOpts(opts.Initialization{Theme: "dark"}),
		charts.WithDataZoomOpts(opts.DataZoom{Type: "slider", Start: 0, End: 100}),
		charts.WithGridOpts(opts.Grid{Left: "10%", Right: "10%", Bottom: "15%", Top: "20%"}),
		charts.WithToolboxOpts(opts.Toolbox{Show: true, Feature: &opts.ToolBoxFeature{
			SaveAsImage: &opts.ToolBoxFeatureSaveAsImage{Show: true, Name: "save", Title: "save"},
			DataZoom:    &opts.ToolBoxFeatureDataZoom{Show: true, Title: map[string]string{"zoom": "zoom", "back": "back"}},
			DataView:    &opts.ToolBoxFeatureDataView{Show: true, Title: "raw", Lang: []string{"raw", "exit", "refresh"}},
		}}),
	)

	line.Validate()
	return line
}

func TopSlowTemplates(stats *Stats) *charts.Kline {
	kline := charts.NewKLine()
	kline.SetCaption(fmt.Sprintf("Chart Shows Top Slow Templates (by time taken) (Top %v)", TopK))

	ids := map[string][]int64{}
	var startTime time.Time = time.Now()
	for _, item := range stats.TemplateStart {
		if item.Time.Before(startTime) {
			startTime = item.Time
		}
	}
	for _, item := range stats.TemplateStart {
		ids[item.ID] = append(ids[item.ID], item.Time.Sub(startTime).Milliseconds())
	}
	for _, item := range stats.TemplateEnd {
		ids[item.ID] = append(ids[item.ID], item.Time.Sub(startTime).Milliseconds())
	}

	type entry struct {
		ID        string
		KlineData opts.KlineData
		start     int64
		end       int64
	}
	data := []entry{}

	for a, b := range ids {
		d := entry{
			ID:        a,
			KlineData: opts.KlineData{Value: []int64{b[0], b[1], b[0], b[1]}}, // open, close, min, max (in our case (open =min) and (close = max)))
			start:     b[0],
			end:       b[1],
		}
		data = append(data, d)
	}

	// sort by most time taken
	sort.Slice(data, func(i, j int) bool {
		return data[i].end-data[i].start > data[j].end-data[j].start
	})

	x := make([]string, 0)
	y := make([]opts.KlineData, 0)
	for _, item := range data[:TopK] {
		x = append(x, item.ID)
		y = append(y, item.KlineData)
	}

	kline.SetXAxis(x).AddSeries("templates", y)
	kline.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: fmt.Sprintf("Nuclei: Top %v Slow Templates", TopK)}),
		charts.WithXAxisOpts(opts.XAxis{
			Type:      "category",
			Show:      true,
			AxisLabel: &opts.AxisLabel{Rotate: 90, Show: true, ShowMinLabel: true, ShowMaxLabel: true, Formatter: opts.FuncOpts(`function (value) { return value; }`)},
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Scale:     true,
			Type:      "value",
			Show:      true,
			AxisLabel: &opts.AxisLabel{Show: true, Formatter: opts.FuncOpts(`function (ms) {  return Math.floor(ms/60000) + 'm' + Math.floor((ms/60000 - Math.floor(ms/60000))*60) + 's'; }`)},
		}),
		charts.WithDataZoomOpts(opts.DataZoom{Type: "slider", Start: 0, End: 100}),
		charts.WithGridOpts(opts.Grid{Left: "10%", Right: "10%", Bottom: "40%", Top: "10%"}),
		charts.WithTooltipOpts(opts.Tooltip{Show: true, Trigger: "item", TriggerOn: "mousemove|click", Enterable: true, Formatter: opts.FuncOpts(`function (params) { return params.name ; }`)}),
		charts.WithToolboxOpts(opts.Toolbox{Show: true, Feature: &opts.ToolBoxFeature{
			SaveAsImage: &opts.ToolBoxFeatureSaveAsImage{Show: true, Name: "save", Title: "save"},
			DataZoom:    &opts.ToolBoxFeatureDataZoom{Show: true, Title: map[string]string{"zoom": "zoom", "back": "back"}},
			DataView:    &opts.ToolBoxFeatureDataView{Show: true, Title: "raw", Lang: []string{"raw", "exit", "refresh"}},
		}}),
	)

	return kline
}

func RequestsVSInterval(stats *Stats, interval time.Duration) *charts.Line {
	line := charts.NewLine()
	line.SetCaption("Chart Shows RPS (Requests Per Second) Over Time")

	// for all template execution end time group them by interval
	// data is already sorted but just in case sort again
	sort.Slice(stats.TemplateEnd, func(i, j int) bool {
		return stats.TemplateEnd[i].Time.Before(stats.TemplateEnd[j].Time)
	})

	data := []opts.LineData{}
	temp := 0
	orig := stats.TemplateEnd[0].Time
	startTime := orig
	xaxisData := []int64{}
	for _, v := range stats.TemplateEnd {
		if v.Time.Sub(startTime) > interval {
			millisec := v.Time.Sub(orig).Milliseconds()
			xaxisData = append(xaxisData, millisec)
			data = append(data, opts.LineData{Value: temp, Name: v.Time.Sub(orig).String()})
			temp = 0
			startTime = v.Time
		}
		temp += 1
	}
	line.SetXAxis(xaxisData)
	line.AddSeries("RPS", data, charts.WithLineChartOpts(opts.LineChart{Smooth: false}), charts.WithLabelOpts(opts.Label{Show: true, Position: "top"}))

	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Nuclei: Template Execution", Subtitle: "Time Interval: " + interval.String()}),
		charts.WithXAxisOpts(opts.XAxis{Name: "Time Intervals", Type: "category", AxisLabel: &opts.AxisLabel{Show: true, ShowMaxLabel: true, Formatter: opts.FuncOpts(`function (date) { return (date/1000)+'s'; }`)}}),
		charts.WithYAxisOpts(opts.YAxis{Name: "RPS Value", Type: "value", Show: true}),
		charts.WithInitializationOpts(opts.Initialization{Theme: "dark"}),
		charts.WithDataZoomOpts(opts.DataZoom{Type: "slider", Start: 0, End: 100}),
		charts.WithGridOpts(opts.Grid{Left: "10%", Right: "10%", Bottom: "15%", Top: "20%"}),
		charts.WithToolboxOpts(opts.Toolbox{Show: true, Feature: &opts.ToolBoxFeature{
			SaveAsImage: &opts.ToolBoxFeatureSaveAsImage{Show: true, Name: "save", Title: "save"},
			DataZoom:    &opts.ToolBoxFeatureDataZoom{Show: true, Title: map[string]string{"zoom": "zoom", "back": "back"}},
			DataView:    &opts.ToolBoxFeatureDataView{Show: true, Title: "raw", Lang: []string{"raw", "exit", "refresh"}},
		}}),
	)

	line.Validate()
	return line
}

func ConcurrencyVsTime(stats *Stats, interval time.Duration) *charts.Line {
	line := charts.NewLine()
	line.SetCaption("Chart Shows Concurrency (Total Workers) Over Time")

	// sort all template start and end times
	dataset := []Item{}
	dataset = append(dataset, stats.TemplateStart...)
	dataset = append(dataset, stats.TemplateEnd...)

	// ascending order
	sort.Slice(dataset, func(i, j int) bool {
		return dataset[i].Time.Before(dataset[j].Time)
	})

	// create array with time interval as x-axis and worker count as y-axis
	// entry is a struct with time and poolsize
	type entry struct {
		Time     time.Duration
		poolsize int
	}
	allEntries := []entry{}

	dataIndex := 0
	maxIndex := len(dataset) - 1
	currEntry := entry{}

	lastTime := dataset[0].Time
	for dataIndex <= maxIndex {
		currTime := dataset[dataIndex].Time
		if currTime.Sub(lastTime) > interval {
			// next batch
			currEntry.Time = interval
			allEntries = append(allEntries, currEntry)
			lastTime = dataset[dataIndex-1].Time
		}
		if dataset[dataIndex].ItemType == ItemStart {
			currEntry.poolsize += 1
		} else {
			currEntry.poolsize -= 1
		}
		dataIndex += 1
	}

	plotData := []opts.LineData{}
	xaxisData := []int64{}
	tempTime := time.Duration(0)
	for _, v := range allEntries {
		tempTime += v.Time
		plotData = append(plotData, opts.LineData{Value: v.poolsize, Name: tempTime.String()})
		xaxisData = append(xaxisData, tempTime.Milliseconds())
	}
	line.SetXAxis(xaxisData)
	line.AddSeries("Concurrency", plotData, charts.WithLineChartOpts(opts.LineChart{Smooth: false}), charts.WithLabelOpts(opts.Label{Show: true, Position: "top"}))

	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Nuclei: WorkerPool", Subtitle: "Time Interval: " + interval.String()}),
		charts.WithXAxisOpts(opts.XAxis{Name: "Time Intervals", Type: "category", AxisLabel: &opts.AxisLabel{Show: true, ShowMaxLabel: true, Formatter: opts.FuncOpts(`function (date) { return (date/1000)+'s'; }`)}}),
		charts.WithYAxisOpts(opts.YAxis{Name: "Total Workers", Type: "value", Show: true}),
		charts.WithInitializationOpts(opts.Initialization{Theme: "dark"}),
		charts.WithDataZoomOpts(opts.DataZoom{Type: "slider", Start: 0, End: 100}),
		charts.WithGridOpts(opts.Grid{Left: "10%", Right: "10%", Bottom: "15%", Top: "20%"}),
		charts.WithToolboxOpts(opts.Toolbox{Show: true, Feature: &opts.ToolBoxFeature{
			SaveAsImage: &opts.ToolBoxFeatureSaveAsImage{Show: true, Name: "save", Title: "save"},
			DataZoom:    &opts.ToolBoxFeatureDataZoom{Show: true, Title: map[string]string{"zoom": "zoom", "back": "back"}},
			DataView:    &opts.ToolBoxFeatureDataView{Show: true, Title: "raw", Lang: []string{"raw", "exit", "refresh"}},
		}}),
	)

	line.Validate()
	return line
}

func getCategoryRequestCount(values []Item) map[string][]Item {
	mx := make(map[string][]Item)
	for _, item := range values {
		mx[item.TemplateType] = append(mx[item.TemplateType], item)
	}
	return mx
}
