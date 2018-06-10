package mpchinachu

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	// "log"
	"net/http"

	mp "github.com/mackerelio/go-mackerel-plugin"
	// "github.com/mackerelio/golib/logging"
)

// var logger = logging.GetLogger("metrics.plugin.chinachu")

type ChinachuPlugin struct {
	Prefix   string
	Target   string
	Tempfile string
}

type status struct {
	ConnectedCount int     `json:"connectedCount"`
	Feature        feature `json:"feature"`
}

type feature struct {
	Previewer    bool
	Streamer     bool
	Filer        bool
	Configurator bool
}

var graphdef = map[string]mp.Graphs{
	"connected_count": mp.Graphs{
		Label: "Chinachu - Connected Count",
		Unit:  "integer",
		Metrics: []mp.Metrics{
			{Name: "ConnectedCount", Label: "Count", Diff: false},
		},
	},
	"feature": mp.Graphs{
		Label: "Chinachu - Feature",
		Unit:  "integer",
		Metrics: []mp.Metrics{
			{Name: "Previewer", Label: "Previewer", Diff: false},
			{Name: "Streamer", Label: "Streamer", Diff: false},
			{Name: "Filer", Label: "Filer", Diff: false},
			{Name: "Configurator", Label: "Configurator", Diff: false},
		},
	},
}

// GetStatus サーバー情報取得
// https://github.com/Chinachu/Chinachu/wiki/REST-API#-status
func GetStatus(host string) (status, error) {
	url := fmt.Sprintf("http://%s/api/status.json", host)

	var s status
	response, err := http.Get(url)

	if err != nil {
		return s, err
	}
	defer response.Body.Close()

	byteArray, _ := ioutil.ReadAll(response.Body)

	if err := json.Unmarshal(byteArray, &s); err != nil {
		return s, err
	}

	return s, err
}

// FetchMetrics interface for mackerelplugin
func (m ChinachuPlugin) FetchMetrics() (map[string]float64, error) {
	stat := make(map[string]float64)

	status, err := GetStatus(m.Target)
	if err != nil {
		return nil, err
	}

	stat["ConnectedCount"] = float64(status.ConnectedCount)

	stat["Previewer"] = float64(Bool2Int(status.Feature.Previewer))
	stat["Streamer"] = float64(Bool2Int(status.Feature.Streamer))
	stat["Filer"] = float64(Bool2Int(status.Feature.Filer))
	stat["Configurator"] = float64(Bool2Int(status.Feature.Configurator))

	return stat, nil
}

// Bool2Int bool -> int 1 or 0
func Bool2Int(x bool) int {
	if x {
		return 1
	}
	return 0
}

// GraphDefinition interface for mackerelplugin
func (m ChinachuPlugin) GraphDefinition() map[string]mp.Graphs {
	return graphdef
}

// MetricKeyPrefix interface for mackerelplugin
func (m ChinachuPlugin) MetricKeyPrefix() string {
	if m.Prefix == "" {
		m.Prefix = "chinachu"
	}
	return m.Prefix
}

func Do() {
	optPrefix := flag.String("matric-key-prefix", "chinachu", "Metric key prefix")
	optHost := flag.String("host", "", "chinachu-wui hostname")
	optPort := flag.String("port", "", "chinachu-wui port")
	optTempfile := flag.String("tempfile", "", "Temp file name")
	flag.Parse()

	var plugin ChinachuPlugin

	plugin.Target = fmt.Sprintf("%s:%s", *optHost, *optPort)
	plugin.Prefix = *optPrefix

	helper := mp.NewMackerelPlugin(plugin)

	if *optTempfile != "" {
		helper.Tempfile = *optTempfile
	} else {
		helper.Tempfile = "/tmp/.mackerel-plugin-chinachu"
	}

	helper.Run()
}
