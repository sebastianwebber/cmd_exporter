package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sebastianwebber/cmdr"
)

var (
	cfg         configFile
	cfgLocation string
)

type configFile struct {
	ListenAddress string `yaml:"listen_address"`
	Port          int    `yaml:"port"`

	MetricsPrefix string            `yaml:"metrics_prefix"`
	HelpMessage   string            `yaml:"help_message"`
	Commands      map[string]string `yaml:"commands"`
}

func init() {
	flag.StringVar(&cfgLocation, "config", "cmd_exporter.yml", "configuration file location")
	flag.Parse()

	dat, err := ioutil.ReadFile(cfgLocation)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}
	err = yaml.Unmarshal(dat, &cfg)
	if err != nil {
		log.Fatalf("Error parsing the yaml file: %v", err)
	}
}

// CMDCollector contains the metrics that we will export
type CMDCollector struct {
	metrics map[string]*prometheus.Desc
}

func newCMDCollector() *CMDCollector {

	newMetrics := make(map[string]*prometheus.Desc)

	for k := range cfg.Commands {

		var newName = fmt.Sprintf("%s_%s", cfg.MetricsPrefix, k)

		log.Printf("Registering '%s'...", newName)

		newDesc := prometheus.NewDesc(
			newName,
			cfg.HelpMessage,
			[]string{"output", "error"},
			nil,
		)
		newMetrics[k] = newDesc
	}

	return &CMDCollector{metrics: newMetrics}
}

//Describe will return the help message
func (collector *CMDCollector) Describe(ch chan<- *prometheus.Desc) {

	//Update this section with the each metric you create for a given collector
	for _, metric := range collector.metrics {
		ch <- metric
	}
}

//Collect will run the commands on OS and export the results
func (collector *CMDCollector) Collect(ch chan<- prometheus.Metric) {

	for k, command := range cfg.Commands {
		var metricValue int
		cmd := cmdr.Parse(command)

		osExec := exec.Command(cmd.Command, cmd.Args...)

		var errMsg string
		stdoutStderr, err := osExec.CombinedOutput()
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				metricValue = exitError.ExitCode()
			}
			errMsg = err.Error()
		}

		if errMsg != "" {
			metricValue = -1
		}

		prettyOutput := strings.TrimSuffix(string(stdoutStderr), "\n")
		log.Printf(
			"KEY: %s, CMD: %s %v - OUT: %s - EXIT: %d - ERROR: %s\n",
			k,
			cmd.Command,
			cmd.Args,
			prettyOutput,
			metricValue,
			errMsg)

		//Write latest value for each metric in the prometheus metric channel.
		ch <- prometheus.MustNewConstMetric(
			collector.metrics[k],
			prometheus.GaugeValue,
			float64(metricValue),
			prettyOutput,
			errMsg)
	}

}

func main() {
	log.Println("CMD Collector")
	cmdCollector := newCMDCollector()
	prometheus.MustRegister(cmdCollector)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(fmt.Sprintf("%s:%d", cfg.ListenAddress, cfg.Port), nil)
}
