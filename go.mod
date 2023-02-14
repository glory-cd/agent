module github.com/glory-cd/agent

go 1.13

require (
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20190717042225-c3de453c63f4 // indirect
	github.com/aws/aws-sdk-go v1.34.0
	github.com/glory-cd/utils v0.0.0-20191025045604-884beaec4a21
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/google/go-cmp v0.3.0
	github.com/gorilla/mux v1.7.3
	github.com/hashicorp/go-cleanhttp v0.5.1
	github.com/jlaffaye/ftp v0.0.0-20190721194432-7cd8b0bcf3fc
	github.com/pkg/errors v0.9.1
	github.com/shirou/gopsutil v2.18.12+incompatible
	github.com/shirou/w32 v0.0.0-20160930032740-bb4de0191aa4 // indirect
	github.com/ulikunitz/xz v0.5.6 // indirect
	golang.org/x/sys v0.0.0-20190712062909-fae7ac547cb7 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
)

replace github.com/glory-cd/agent => ./
