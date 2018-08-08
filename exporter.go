// +Version   = "0"
// +Revision  = "08/08/2018"
// +Branch    = "master"
// +BuildUser = "hbermu"
// +BuildDate = "08/08/2018"

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
	case "XVII":
		return "17"
	case "XVIII":
		return "18"
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
	ipString := strings.Trim(strings.Replace(fmt.Sprint(*ips), " ", " ", -1), "[]")
	log.Infoln("Storage MD IPs:", ipString)
	log.Debugln("Transform binary path pointer to string")
	binPathString := *binPath
	log.Infoln("Binary path:", binPathString)

	// Prepare Command
	log.Debugln("Prepare cmd command to launch")
	var cmd *exec.Cmd
	switch command{
	case 0:
		cmd = exec.Command("/bin/sh", binPathString, ipString, "-S", "-c 'show allphysicaldisks performancestats;'" )
		log.Debugf("Launch command: /bin/sh %v %v -S -c 'show allphysicaldisks performancestats;'", binPathString, ipString)
	case 1:
		cmd = exec.Command("/bin/sh", binPathString, ipString, "-S", "-c 'show allvirtualdisks performancestats;'" )
		log.Debugf("Launch command: /bin/sh %v %v -S -c 'show allvirtualdisks performancestats;'", binPathString, ipString)
	default:
		return nil; log.Fatal("Wrong Command Index")
	}

	// Create buffer
	var out bytes.Buffer
	cmd.Stdout = &out

	// Run command and take output
	log.Debugln("Running command")
	//out, err := cmd.Output()
	err := cmd.Run()
	if err != nil {
		log.Fatal("Error to execute command: ", err)
	}

	//////////////////
	// Parse output //
	//////////////////
	log.Debugln("Parse command output to string")
	// Output -> []String
	//outString := out.String()

	outString := strings.SplitN(out.String(),  "\n", 2)[1]
	//log.Debugln("Output command:\n", strings.SplitN(outString,  "\n", 2)[1])
	log.Debugln("Output command:\n", outString )

	// Create reader CSV
	log.Debugln("Creating reader CSV")
	reader := csv.NewReader(strings.NewReader(outString))
	// Save all records to array String
	log.Debugln("Saving all CSV record to [][]String")

	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	log.Debugln("CSV: \n", records)


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
		log.Debugln("Object:", object[0])
		if object[0] == "Expansion" {

			// Take parameters witch we want
			log.Infoln("Taking parameters for object", object)
			enclosure 	:= strings.Replace(object[2], ",", "", -1)
			drawer 		:= strings.Replace(object[4], ",", "", -1)
			slot 		:= strings.Replace(object[6], ",", "", -1)
			log.Debugf("Enclosure: %v, Drawer: %v, Slot: %v", enclosure, drawer, slot)

			// Parse value to float 64
			log.Debugln("Parsing value to float 64")
			value, err := strconv.ParseFloat(element[1], 64)
			if err != nil {
				log.Fatal(err)
			}
			log.Debugln("Value: ", value)

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
		log.Debugln("Object:", object[0])
		if object[0] == "Virtual" {

			log.Infoln("Taking parameters for object", object)
			log.Debugln("Split Disk Name")
			disk := strings.Split(object[2], "_")

			// Take parameters witch we want
			log.Infoln("Identify Disk type and number")
			typeDisk 	:= disk[1]
			numberDisk 	:= romToDec(disk[2])
			log.Debugf("Disk Type: %v, Number: %v", typeDisk, numberDisk)

			// Parse value to float 64
			log.Debugln("Parsing values to float 64")
				// Latency
			latency, err := strconv.ParseFloat(element[14], 64)
			if err != nil {
				log.Fatal(err)
			}
			log.Debugln("Latency:", latency)
				// IO
			ioCurrent, err := strconv.ParseFloat(element[8], 64)
			if err != nil {
				log.Fatal(err)
			}
			log.Debugln("IO Current:", ioCurrent)
				// Speed
			speed, err := strconv.ParseFloat(element[6], 64)
			if err != nil {
				log.Fatal(err)
			}
			log.Debugln("Speed:", speed)

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

	var (
		listenAddress	= kingpin.Flag("listen-address", "Address to listen on for web interface and telemetry.").Default(":9362").String()
		metricsPath		= kingpin.Flag("telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
		cabinesIPs		= kingpin.Flag("IP", "IPs Dell Cabine Storage. Example: '--IP=172.19.0.1 --IP=172.0.0.1'").Default("172.0.0.1").IPList()
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