module github.com/summerwind/protospec

go 1.16

require (
	github.com/deislabs/oras v0.8.1
	github.com/ghodss/yaml v1.0.0
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/opencontainers/image-spec v1.0.1
	github.com/prometheus/common v0.14.0 // indirect
	github.com/prometheus/procfs v0.2.0 // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.1.1
	golang.org/x/net v0.0.0-20200927032502-5d4f70055728
	golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208 // indirect
	golang.org/x/text v0.3.3 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gotest.tools/v3 v3.0.3 // indirect
)

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.4.3
	golang.org/x/sys => golang.org/x/sys v0.0.0-20200826173525-f9321e4c35a6
)
