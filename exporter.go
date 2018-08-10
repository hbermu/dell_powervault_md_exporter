// +Version   = "0.0.1"
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

	"time"
	"runtime"
)

// Const
const (
	namespace 	= "dell_powervault_md_exporter" // For Prometheus metrics.
	Version   	= "0.0.1"
	Revision  	= "09/08/2018"
	Branch    	= "master"
	//BuildUser 	= "hbermu"
	//BuildDate 	= "09/08/2018"
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
	case "IXX":
		return "19"
	case "XX":
		return "20"
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

var (
	// Register metrics for Physical Disks Performance
	metricPhysicalDisksLatency = newMetric(
		"Physical_Disk_Latency",
		"Delay from input into a system to desired outcome.",
		[]string{"Enclosure", "Drawer", "Slot"})
	metricPhysicalDisksStatus = newMetric(
		"Physical_Disk_Status",
		"Status physical disks. 1: OK, 0: Something Wrong",
		[]string{"Enclosure", "Drawer", "Slot"})
	// Register metrics for Virtual Disks Performance
	metricVirtualDisksLatency = newMetric(
		"Virtual_Disk_Latency",
		"Delay from input into a system to desired outcome.",
		[]string{"Type", "Index"})
	metricVirtualDisksIO = newMetric(
		"Virtual_Disk_IO",
		"Input/Output operations on a physical disk.",
		[]string{"Type", "Index"})
	metricVirtualDisksSpeed = newMetric(
		"Virtual_Disk_Speed",
		"Speed at the disk is rotates.",
		[]string{"Type", "Index"})

)

func init() {
	// Register the summary and the histogram with Prometheus's default registry.
		// Physical
	prometheus.MustRegister(metricPhysicalDisksLatency)
	prometheus.MustRegister(metricPhysicalDisksStatus)
		// Virtual
	prometheus.MustRegister(metricVirtualDisksLatency)
	prometheus.MustRegister(metricVirtualDisksIO)
	prometheus.MustRegister(metricVirtualDisksSpeed)

	// Build Info Program
	buildInfo := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "build_info",
			Help: fmt.Sprintf(
				"A metric with a constant '1' value labeled by version, revision, branch, and goversion from which %s was built.",
				namespace,
			),
		},
		[]string{"version", "revision", "branch", "goversion"},
	)
	buildInfo.WithLabelValues(Version, Revision, Branch, runtime.Version()).Set(1)

}

func getRecords (binPath *string, ips *[]net.IP, command int) string{

	// Array IPs -> String
	log.Debugln("Transform Array IPs pointer to string")
	ipString := strings.Trim(strings.Replace(fmt.Sprint(*ips), " ", " ", -1), "[]")
	log.Debugln("Transform binary path pointer to string")
	binPathString := *binPath

	// Prepare Command
	log.Debugln("Prepare cmd command to launch")
	var cmd *exec.Cmd
	switch command{
	case 0:
		cmd = exec.Command(binPathString, ipString, "-S", "-c", "show allphysicaldisks performancestats;")
		log.Debugf("Launch command: %v %v -S -c 'show allphysicaldisks performancestats;'", binPathString, ipString)
	case 1:
		cmd = exec.Command(binPathString, ipString, "-S", "-c", "show allvirtualdisks performancestats;" )
		log.Debugf("Launch command: %v %v -S -c 'show allvirtualdisks performancestats;'", binPathString, ipString)
	case 2:
		cmd = exec.Command(binPathString, ipString, "-S", "-c", "show allphysicaldisks summary;" )
		log.Debugf("Launch command: %v %v -S -c 'show allphysicaldisks summary;'", binPathString, ipString)
	default:
		return ""; log.Fatal("Wrong Command Index")
	}

	// Create output buffer
	var out bytes.Buffer
	cmd.Stdout = &out

	// Run command and take output
	log.Debugln("Running command")
	err := cmd.Run()
	if err != nil {
		log.Warn("Error to execute command: ", err)
		log.Warn("Output command:\n", out.String())
		return ""
	}

	// Output -> String
	log.Debugln("Parse command output to string")
	outString := out.String()
	log.Debugln("Output command:\n", outString )

	return outString

}

func parseRecords(out string) [][]string{

	log.Debugln("Parse command output to [][]string CSV")
	// The first line isn't CVS format
	outString := strings.SplitN(out,  "\n", 2)[1]
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

	// Get all records
	log.Infoln("Getting all Physical records")
	output := getRecords(binPath, ips, 0)
	if strings.EqualFold(output, "") {
		log.Warn("Waiting to repeat the petition")
		return
	}
	records := parseRecords(output)

	// For each row
	for _, element := range records {

		// Split the first element to take only the object we want
		log.Debugln("Physical -- Split first string")
		object := strings.Split(element[0], " ")
		log.Debugln("Physical -- Object:", object[0])
		if object[0] == "Expansion" {

			// Take parameters witch we want
			log.Debugln("Physical -- Taking parameters for object", object)
			enclosure 	:= strings.Replace(object[2], ",", "", -1)
			drawer 		:= strings.Replace(object[4], ",", "", -1)
			slot 		:= strings.Replace(object[6], ",", "", -1)
			log.Debugf("Physical -- Enclosure: %v, Drawer: %v, Slot: %v", enclosure, drawer, slot)

			// Parse value to float 64
			log.Debugln("Physical -- Parsing value to float 64")
			value, err := strconv.ParseFloat(element[1], 64)
			if err != nil {
				log.Fatal(err)
			}
			log.Debugln("Physical -- Value: ", value)

			// Add value to the metrics vector
			log.Debugln("Physical -- Adding value to Vector Metric")
			metricPhysicalDisksLatency.WithLabelValues(
				enclosure,
				drawer,
				slot).Set(value)
		}
	} // End for
	log.Infoln("Getting all Physical records done")
}

func virtualDisksPerformance(binPath *string, ips *[]net.IP) {

	// Get all records
	log.Infoln("Getting all Virtual records")
	output := getRecords(binPath, ips, 0)
	if strings.EqualFold(output, "") {
		log.Warn("Virtual -- Waiting to repeat the petition")
		return
	}
	records := parseRecords(getRecords(binPath, ips, 1))

	// For each row
	for _, element := range records {

		// Split the first element to take only the object we want
		log.Debugln("Virtual -- Split first string")
		object := strings.Split(element[0], " ")
		log.Debugln("Virtual -- Object:", object[0])
		if object[0] == "Virtual" {

			log.Debugln("Virtual -- Taking parameters for object", object)
			log.Debugln("Virtual -- Split Disk Name")
			disk := strings.Split(object[2], "_")

			// Take parameters witch we want
			log.Debugln("Virtual -- Identify Disk type and number")
			typeDisk 	:= disk[1]
			numberDisk 	:= romToDec(disk[2])
			log.Debugf("Virtual -- Disk Type: %v, Number: %v", typeDisk, numberDisk)

			// Parse value to float 64
			log.Debugln("Virtual -- Parsing values to float 64")
				// Latency
			latency, err := strconv.ParseFloat(element[14], 64)
			if err != nil {
				log.Fatal(err)
			}
			log.Debugln("Virtual -- Latency:", latency)
				// IO
			ioCurrent, err := strconv.ParseFloat(element[8], 64)
			if err != nil {
				log.Fatal(err)
			}
			log.Debugln("Virtual -- IO Current:", ioCurrent)
				// Speed
			speed, err := strconv.ParseFloat(element[6], 64)
			if err != nil {
				log.Fatal(err)
			}
			log.Debugln("Virtual -- Speed:", speed)

			// Add value to the metrics vector
			log.Debugln("Virtual -- Adding values to Vector Metric")
				// Latency
			metricVirtualDisksLatency.WithLabelValues(
				typeDisk,
				numberDisk).Set(latency)
				// IO
			metricVirtualDisksIO.WithLabelValues(
				typeDisk,
				numberDisk).Set(ioCurrent)
				// Speed
			metricVirtualDisksSpeed.WithLabelValues(
				typeDisk,
				numberDisk).Set(speed)
		}
	} // End for
	log.Infoln("Getting all Virtual records done")
}

func physicalDisksSummary(binPath *string, ips *[]net.IP) {

	// Get all records
	log.Infoln("Getting Physical summary")
	records := getRecords(binPath, ips, 2)
	if strings.EqualFold(records, "") {
		log.Warn("Physical Status -- Waiting to repeat the petition")
		return
	}

	// Split lines
	log.Debugln("Physical Status -- Split all string")
	recordsArray := strings.Split(records, "\n")

	// Take number of disks
	numberDisks, err := strconv.Atoi(strings.Split(recordsArray[1], " ")[7])
	if err != nil {
		log.Fatal(err)
	}
	log.Debugln("Physical Status -- Disks: ", numberDisks)

	// Take all disks
	for i := 8; i < (numberDisks + 8); i++ {
		log.Debugln("Physical Status -- Parsing Row")
		// Split for blank spaces " "
		row := strings.Split(recordsArray[i], " ")
		// Clean all "" positions
		var rowClean []string
		for _, str := range row {
			if str != "" {
				rowClean = append(rowClean, str)
			}
		}
		log.Debugln("Physical Status -- Row: ", rowClean)

		// Get Disk position
		enclosure 	:= strings.Replace(rowClean[0], ",", "", -1)
		drawer 		:= strings.Replace(rowClean[1], ",", "", -1)
		slot 		:= rowClean[2]
		log.Debugf("Physical Status -- Enclosure: %v, Drawer: %v, Slot: %v", enclosure, drawer, slot)

		// Define Disk Status
			// Prepare for update
		var value float64
		switch rowClean[3]{
		case "Optimal":
			value = 1
		default:
			value = -1
		}
		log.Debugln("Physical Status -- Value: ", value)

		log.Debugln("Physical Status -- Adding values to Vector Metric")
		// Latency
		metricPhysicalDisksStatus.WithLabelValues(
			enclosure,
			drawer,
			slot).Set(value)
	}
	log.Infoln("Getting all Physical Status done")
}


func main() {

	var (
		listenAddress	= kingpin.Flag("port", "Port to listen on for web interface and telemetry.").Short('p').Default(":9362").String()
		metricsPath		= kingpin.Flag("telemetry-path", "Path under which to expose metrics.").Short('m').Default("/metrics").String()
		cabinesIPs		= kingpin.Flag("IP", "IPs Dell Cabine Storage. Example: '--IP=172.19.0.1 --IP=172.0.0.1'").Default("172.0.0.1").IPList()
		sMcliPath		= kingpin.Flag("SMcliPath", "SMcli binary path").Default("/opt/dell/mdstoragesoftware/mdstoragemanager/client/SMcli").String()
		timeEach		= kingpin.Flag("time", "Time between petitions").Default("30s").Duration()

	)

	// Check flags
	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("dell_powervault_md_exporter"))
	kingpin.HelpFlag.Short('d')
	kingpin.Parse()

	// Log Start
	log.Infoln("Starting dell_powervault_md_exporterr", version.Info())
	log.Infoln("Build context", version.BuildContext())
	log.Infoln("Storage MD IPs:", *cabinesIPs)
	log.Infoln("Binary path:", *sMcliPath)

	// Transform time
	delay := *timeEach

	go func() {
		for {
			physicalDisksPerformance(sMcliPath, cabinesIPs)
			virtualDisksPerformance(sMcliPath, cabinesIPs)
			physicalDisksSummary(sMcliPath, cabinesIPs)
			log.Debugf("Waiting %v until next petitions", timeEach.String())
			time.Sleep(delay)
		}
	}()

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