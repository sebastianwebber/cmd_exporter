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

	MetricName  string `yaml:"metric_name"`
	HelpMessage string `yaml:"help_message"`
	Command     string `yaml:"command"`
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
	cmdMetric *prometheus.Desc
}

func newCMDCollector() *CMDCollector {
	return &CMDCollector{
		cmdMetric: prometheus.NewDesc(cfg.MetricName,
			cfg.HelpMessage,
			[]string{"output"},
			nil,
		),
	}
}

//Describe will return the help message
func (collector *CMDCollector) Describe(ch chan<- *prometheus.Desc) {

	//Update this section with the each metric you create for a given collector
	ch <- collector.cmdMetric
}

//Collect will run the commands on OS and export the results
func (collector *CMDCollector) Collect(ch chan<- prometheus.Metric) {
	var metricValue int
	cmd := cmdr.Parse(cfg.Command)

	osExec := exec.Command(cmd.Command, cmd.Args...)
	stdoutStderr, err := osExec.CombinedOutput()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			metricValue = exitError.ExitCode()
		}
	}

	prettyOutput := strings.TrimSuffix(string(stdoutStderr), "\n")
	log.Printf("CMD: %s %v - OUT: %s - EXIT: %d\n", cmd.Command, cmd.Args, prettyOutput, metricValue)

	//Write latest value for each metric in the prometheus metric channel.
	ch <- prometheus.MustNewConstMetric(
		collector.cmdMetric,
		prometheus.GaugeValue,
		float64(metricValue),
		fmt.Sprintf("output=%s", prettyOutput))

}

func main() {
	log.Println("CMD Collector")
	cmdCollector := newCMDCollector()
	prometheus.MustRegister(cmdCollector)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(fmt.Sprintf("%s:%d", cfg.ListenAddress, cfg.Port), nil)
}
