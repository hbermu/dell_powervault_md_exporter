package main

import (
	"encoding/csv"
	"fmt"
		"net"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"strings"
	"os/exec"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
	"bytes"
)

// Const
const (
	namespace = "dell_powervault_md_exporter" // For Prometheus metrics.
)

func romToDec (number string) string{
	switch number {
	case "I":
		return "1"
	case "II":
		return "2"
	case "III":
		return "3"
	case "IV":
		return "4"
	case "V":
		return "5"
	case "VI":
		return "6"
	case "VII":
		return "7"
	case "VIII":
		return "8"
	case "IX":
		return "9"
	case "X":
		return "10"
	case "XI":
		return "11"
	case "XII":
		return "12"
	case "XIII":
		return "13"
	case "XIV":
		return "14"
	case "XV":
		return "15"
	case "XVI":
		return "16"
	}

	log.Fatal("Wrong Roman Number")
	return ""

}

func newMetric(metricName string, docString string, labelNames []string) *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Name:        metricName,
			Help:        docString,
		},
		labelNames,
	)
}

func getRecords (binPath *string, ips *[]net.IP, command int) [][]string{

	// Array IPs -> String
	log.Debugln("Transform Array IPs pointer to string")
	ipString := strings.Trim(strings.Replace(fmt.Sprint(ips), " ", " ", -1), "[]")
	log.Infoln("Storage MD IPs: %s", ipString)
	log.Debugln("Transform binary path pointer to string")
	binPathString := *binPath
	log.Infoln("Binary path: %s", binPathString)

	// Prepare Command
	log.Debugln("Prepare cmd command to launch")
	var cmd *exec.Cmd
	switch command{
	case 0:
		cmd = exec.Command(binPathString, ipString, "-S", "-c 'show allphysicaldisks performancestats;'" )
		log.Infoln("Command to launch: %s %s -S -c 'show allphysicaldisks performancestats;'")
	case 1:
		cmd = exec.Command(binPathString, ipString, "-S", "-c 'show allvirtualdisks performancestats;'" )
		log.Infoln("Command to launch: %s %s -S -c 'show allvirtualdisks performancestats;'")
	default:
		return nil; log.Fatal("Wrong Command Index")
	}

	// Run command and take output
	log.Debugln("Running command")
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	//////////////////
	// Parse output //
	//////////////////
	log.Debugln("Parse command output")
	// Output length
	outLength := bytes.Index(out, []byte{0})
	log.Debugln("Output length %v", outLength)
	// Output bytes -> String
	outString := string(out[:outLength])
	log.Infoln("Output command: %s", outString)

	// Create reader CSV
	log.Debugln("Creating reader CSV")
	reader := csv.NewReader(strings.NewReader(outString))
	// Save all records to array String
	log.Debugln("Saving all CSV record to [][]String")
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	// And returns them
	return records

}

func physicalDisksPerformance(binPath *string, ips *[]net.IP) {

	// Register new metric type
	log.Infoln("Register metrics for Physical Disks Perfomance")
	metricPhysicalDisksLatency := newMetric(
		"Physical_Disk_Latency",
		"Delay from input into a system to desired outcome.",
		[]string{"Enclosure", "Drawer", "Slot"})
	prometheus.MustRegister(metricPhysicalDisksLatency)

	// Get all records
	log.Infoln("Getting al records")
	records := getRecords(binPath, ips, 0)

	// For each row
	log.Infoln("Analyze all records")
	for _, element := range records {

		// Split the first element to take only the object we want
		log.Debugln("Split first string")
		object := strings.Split(element[0], " ")
		log.Debugln("Object: %s", object[0])
		if object[0] == "Expansion" {

			// Take parameters witch we want
			log.Infoln("Taking parameters for object %s", object[0])
			enclosure 	:= strings.Replace(object[2], ",", "", -1)
			drawer 		:= strings.Replace(object[4], ",", "", -1)
			slot 		:= strings.Replace(object[6], ",", "", -1)
			log.Debugln("Enclosure: %s, Drawer: %s, Slot: %s", enclosure, drawer, slot)

			// Parse value to float 64
			log.Debugln("Parsing value to float 64")
			value, err := strconv.ParseFloat(element[1], 64)
			if err != nil {
				log.Fatal(err)
			}
			log.Debugln("Value %v", value)

			// Add value to the metrics vector
			log.Infoln("Adding value to Vector Metric")
			metricPhysicalDisksLatency.WithLabelValues(
				enclosure,
				drawer,
				slot).Add(value)
		}
	} // End for


}

func virtualDisksPerformance(binPath *string, ips *[]net.IP) {

	// Register new metric type
	log.Infoln("Register metrics for Physical Disks Perfomance")
	metricVirtualDisksLatency := newMetric(
		"Virtual_Disk_Latency",
		"Delay from input into a system to desired outcome.",
		[]string{"Type", "Index"})
	metricVirtualDisksIO := newMetric(
		"Virtual_Disk_IO",
		"Input/Output operations on a physical disk.",
		[]string{"Type", "Index"})
	metricVirtualDisksSpeed := newMetric(
		"Virtual_Disk_Speed",
		"Speed at the disk is rotates.",
		[]string{"Type", "Index"})
	prometheus.MustRegister(metricVirtualDisksIO)
	prometheus.MustRegister(metricVirtualDisksLatency)
	prometheus.MustRegister(metricVirtualDisksSpeed)

	// Get all records
	log.Infoln("Getting al records")
	records := getRecords(binPath, ips, 1)

	// For each row
	log.Infoln("Analyze all records")
	for _, element := range records {

		// Split the first element to take only the object we want
		log.Debugln("Split first string")
		object := strings.Split(element[0], " ")
		log.Debugln("Object: %s", object[0])
		if object[0] == "Virtual" {

			log.Debugln("Split Disk Name")
			disk := strings.Split(object[2], "_")

			// Take parameters witch we want
			log.Infoln("Identify Disk type and number")
			typeDisk 	:= disk[1]
			numberDisk 	:= romToDec(disk[2])
			log.Debugln("Disk Type: %s, Number: %s", typeDisk, numberDisk)

			// Parse value to float 64
			log.Debugln("Parsing values to float 64")
				// Latency
			latency, err := strconv.ParseFloat(element[14], 64)
			if err != nil {
				log.Fatal(err)
			}
			log.Debugln("Latency: %v", latency)
				// IO
			ioCurrent, err := strconv.ParseFloat(element[8], 64)
			if err != nil {
				log.Fatal(err)
			}
			log.Debugln("IO Current: %v", ioCurrent)
				// Speed
			speed, err := strconv.ParseFloat(element[6], 64)
			if err != nil {
				log.Fatal(err)
			}
			log.Debugln("Speed: %v", speed)

			// Add value to the metrics vector
			log.Infoln("Adding values to Vector Metric")
				// Latency
			metricVirtualDisksLatency.WithLabelValues(
				typeDisk,
				numberDisk).Add(latency)
				// IO
			metricVirtualDisksLatency.WithLabelValues(
				typeDisk,
				numberDisk).Add(ioCurrent)
				// Speed
			metricVirtualDisksLatency.WithLabelValues(
				typeDisk,
				numberDisk).Add(speed)
		}
	} // End for

}


func main() {

	const pidFileHelpText = `Path to DellDiskStorage pid file.

	If provided, the standard process metrics get exported for the DellDiskStorage
	process.

	https://prometheus.io/docs/instrumenting/writing_clientlibs/#process-metrics.`

	var (
		listenAddress	= kingpin.Flag("listen-address", "Address to listen on for web interface and telemetry.").Default(":9362").String()
		metricsPath		= kingpin.Flag("telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
		cabinesIPs		= kingpin.Flag("IPs", "IPs Dell Cabine Storage. Example: '172.19.0.1 172.0.0.1'").Default("172.0.0.1").IPList()
		sMcliPath		= kingpin.Flag("SMcliPath", "SMcli binary path").Default("/opt/dell/mdstoragesoftware/mdstoragemanager/client/SMcli").String()

	)

	// Check flags
	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("dell_powervault_md_exporterr"))
	kingpin.HelpFlag.Short('d')
	kingpin.Parse()

	// Log Start
	log.Infoln("Starting dell_powervault_md_exporterr", version.Info())
	log.Infoln("Build context", version.BuildContext())

	// Take all metrics
	physicalDisksPerformance(sMcliPath, cabinesIPs)
	virtualDisksPerformance(sMcliPath, cabinesIPs)

	// Start listen
	log.Infoln("Listening on", *listenAddress)
	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Dell PowerVault MD Exporter</title></head>
             <body>
             <h1>Dell PowerVault MD Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}