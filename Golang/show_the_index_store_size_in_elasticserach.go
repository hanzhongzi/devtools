package main

import (
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

/*
var (
	bar3DRangeColor = []string{
		"#313695", "#4575b4", "#74add1", "#abd9e9", "#e0f3f8",
		"#fee090", "#fdae61", "#f46d43", "#d73027", "#a50026",
	}

	bar3DHrs = [...]string{
		"12a", "1a", "2a", "3a", "4a", "5a", "6a", "7a", "8a", "9a", "10a", "11a",
		"12p", "1p", "2p", "3p", "4p", "5p", "6p", "7p", "8p", "9p", "10p", "11p",
	}

	bar3DDays = [...]string{"Saturday", "Friday", "Thursday", "Wednesday", "Tuesday", "Monday", "Sunday"}
)*/

func mapToSlice(m map[string]string) []string {
	s := make([]string, 0, len(m))
	for _, v := range m {
		s = append(s, v)
	}
	return s
}

func mapToSliceInt(m map[int64]int64) []int64 {
	s := make([]int64, 0, len(m))
	for _, v := range m {
		s = append(s, v)
	}
	return s
}

func main() {
	var metrics_messages *io_prometheus_client.MetricFamily

	resp, err := http.Get("http://10.134.32.251:9114/metrics")
	if err != nil {
		defer fmt.Println(err, "http request error!")
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		defer fmt.Println(err, "http request error!")
		return
	}
	// fmt.Println(string(body))
	var parser expfmt.TextParser
	messages, err := parser.TextToMetricFamilies(strings.NewReader(string(body)))
	if err != nil {
		defer fmt.Println(err, "http request error!")
		return
	}
	for k, v := range messages {
		if k == "elasticsearch_indices_store_size_bytes_total" {
			metrics_messages = v
		}
	}
	//fmt.Println("Name", *metrics_messages.Name)
	//fmt.Println("Help", *metrics_messages.Help)
	//fmt.Println("Type", metrics_messages.Type)
	// fmt.Println("Metric", metrics_messages.Metric)
	//fmt.Println("Name", *metrics_messages.Name)
	elasticsearch_index := map[string][]string{}
	for _, v := range metrics_messages.Metric {
		if v.GetLabel()[1].GetValue()[0] == '.' {
			continue
		}
		if !strings.Contains(v.GetLabel()[1].GetValue(), "2022") {
			continue
		}
		cluster := v.GetLabel()[0].GetValue()
		index := v.GetLabel()[1].GetValue()
		//fmt.Print(v.GetLabel()[1].GetValue(), ": ")
		indexprestring := strings.Split(v.GetLabel()[1].GetValue(), "2022")[0]
		indexdate := "2022" + strings.Split(v.GetLabel()[1].GetValue(), "2022")[1]
		//values := strconv.FormatInt(int64(v.GetGauge().GetValue()/1024/1024), 10) //MB
		values := strconv.FormatInt(int64(v.GetGauge().GetValue()), 10) //Bytes
		fmt.Println(cluster, index, indexprestring, indexdate, values)
		elasticsearch_index[index] = []string{cluster, indexprestring, indexdate, index, values}
		// 关系[索引]={cluster,bar3D_X,bar3D_Y,index_name,bar3D_Z}
	}

	//========
	//size := len(elasticsearch_index)
	bar3DRangeColor := []string{
		"#313695", "#4575b4", "#74add1", "#abd9e9", "#e0f3f8",
		"#fee090", "#fdae61", "#f46d43", "#d73027", "#a50026",
	}

	bar3D_X := make(map[string]string)
	for _, v := range elasticsearch_index {
		bar3D_X[v[1]] = v[1]
	}
	bar3D_X_slice := mapToSlice(bar3D_X)
	sort.Strings(bar3D_X_slice)
	//fmt.Println(bar3D_X_slice)
	bar3D_Y := make(map[string]string)
	for _, v := range elasticsearch_index {
		bar3D_Y[v[2]] = v[2]
	}
	bar3D_Y_slice := mapToSlice(bar3D_Y)
	sort.Strings(bar3D_Y_slice)
	//fmt.Println(bar3D_Y_slice)
	bar3D_Z := make(map[int64]int64)
	for _, v := range elasticsearch_index {
		v, _ := strconv.ParseInt(v[4], 10, 64)
		bar3D_Z[v] = v
	}
	bar3D_Z_slice := mapToSliceInt(bar3D_Z)
	//fmt.Println(bar3D_Z_slice)

	bar3D_data := make([][]int, 0, len(bar3D_Z_slice))

	for _, v := range bar3D_Z_slice {
		for index_pre_i, index_pre_v := range bar3D_X_slice {
			for date_i, date_v := range bar3D_Y_slice {

				if len(elasticsearch_index[index_pre_v+date_v]) > 0 {

					v_value, _ := strconv.ParseInt(elasticsearch_index[index_pre_v+date_v][4], 10, 64)

					if v_value == v {
						//fmt.Println(bar3D_Z_slice[i], i)
						//fmt.Println(bar3D_X_slice[index_pre_i], index_pre_i)
						//fmt.Println(bar3D_Y_slice[date_i], date_i)
						//fmt.Println(index_pre_v+date_v, "----", elasticsearch_index[index_pre_v+date_v])
						bar3D_data = append(bar3D_data, []int{index_pre_i, date_i, int(v)})

					}
				}

			}
		}
	}

	//fmt.Println(bar3D_data)
	exp := Bar3dExamples{}
	exp.Examples(bar3DRangeColor, bar3D_X_slice, bar3D_Y_slice, bar3D_data)
}
func genBar3dData(bar3DDays [][]int) []opts.Chart3DData {
	for i := 0; i < len(bar3DDays); i++ {
		bar3DDays[i][0], bar3DDays[i][1] = bar3DDays[i][1], bar3DDays[i][0]
	}
	for i := 0; i < len(bar3DDays); i++ {
		bar3DDays[i][0], bar3DDays[i][1] = bar3DDays[i][1], bar3DDays[i][0]
	}

	ret := make([]opts.Chart3DData, 0)
	for _, d := range bar3DDays {
		ret = append(ret, opts.Chart3DData{
			Value: []interface{}{d[0], d[1], d[2]},
		})
	}

	return ret
}

func bar3DAutoRotate(bar3DRangeColor []string, bar3D_X_slice []string, bar3D_Y_slice []string, bar3DDays [][]int) *charts.Bar3D {
	bar3d := charts.NewBar3D()
	bar3d.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "auto rotating"}),
		charts.WithParallelAxisList([]opts.ParallelAxis{{Inverse: false}}),
		charts.WithVisualMapOpts(opts.VisualMap{
			Calculable: true,
			Max:        200000000000,
			Range:      []float32{0, 200000000000},
			InRange:    &opts.VisualMapInRange{Color: bar3DRangeColor},
		}),
		charts.WithGrid3DOpts(opts.Grid3D{
			BoxWidth: 200,
			BoxDepth: 300,
			//ViewControl: &opts.ViewControl{AutoRotate: true},
		}),
		charts.WithInitializationOpts(opts.Initialization{Width: "1500px", Height: "900px"}),
	)

	bar3d.SetGlobalOptions(
		charts.WithXAxis3DOpts(opts.XAxis3D{Data: bar3D_X_slice, Name: "索引名称"}),
		charts.WithYAxis3DOpts(opts.YAxis3D{Data: bar3D_Y_slice, Name: "索引生成日期"}),
	)
	bar3d.AddSeries("bar3d", genBar3dData(bar3DDays))
	return bar3d
}

type Bar3dExamples struct{}

func (Bar3dExamples) Examples(bar3DRangeColor []string, bar3D_X_slice []string, bar3D_Y_slice []string, bar3DDays [][]int) {
	page := components.NewPage()
	page.Height = "720"
	page.Width = "1080"
	page.AddCharts(
		bar3DAutoRotate(bar3DRangeColor, bar3D_X_slice, bar3D_Y_slice, bar3DDays),
	)
	f, err := os.Create("bar3d.html")
	if err != nil {
		panic(err)
	}
	page.Render(io.MultiWriter(f))
}

