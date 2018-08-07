# Dell Disk Storage Exporter

Prometheus exporter for Dell PowerVault MD32/MD34/MD36/MD38 Series, in Go with pluggable metric collectors.

## Building and running
Prerequisites:

* [Go compiler](https://golang.org/dl/)
* [SMcli tool](https://www.dell.com/support/home/en/en/esbsdt1/drivers/driversdetails?driverid=jtpc2)

Supported Operating Systems:
* Red Hat Enterprise Linux 6
* Red Hat Enterprise Linux 7
* Suse Linux ES 10

Building:
```
go get github.com/hbermu/dellDiskStorage-Exporter
cd ${GOPATH-$HOME/go}/src/github.com/hbermu/dellDiskStorage-Exporter
go build -o dellDiskStorage-Exporter .
./dellDiskStorage-Exporter <flags>
```

To see all available configuration flags:
```
./node_exporter -h
```

