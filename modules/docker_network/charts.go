package docker_network

import (
	"fmt"
	"github.com/netdata/go.d.plugin/agent/module"
	"strings"
)

const (
	prioNetworkBytes = module.Priority + iota
)

var summaryCharts = module.Charts{
	// These charts will be collected for all of docker
	// Mostly just aggregate stats
	networkBytesChart.Copy(),
}

var (
	networkBytesChart = module.Chart{
		ID:       "network_bytes",
		Title:    "Network bytes",
		Units:    "bytes/s",
		Fam:      "network",
		Ctx:      "docker_net.network_bytes",
		Priority: prioNetworkBytes,
		Type:     module.Stacked,
		Dims: module.Dims{
			{ID: "network_bytes_rx", Name: "received"},
			{ID: "network_bytes_tx", Name: "sent"},
		},
	}
)

var (
	containerNetworkChartsTmpl = module.Charts{
		// These will be collected for each container
		containerNetworkBytesChartTmpl.Copy(),
	}

	containerNetworkBytesChartTmpl = module.Chart{
		ID:       "network_%s_bytes",
		Title:    "Network bytes",
		Units:    "bytes/s",
		Fam:      "network",
		Ctx:      "docker_net.container_network_bytes",
		Priority: prioNetworkBytes,
		Type:     module.Stacked,
		Dims: module.Dims{
			{ID: "network_%s_bytes_rx", Name: "received"},
			{ID: "network_%s_bytes_tx", Name: "sent"},
		},
	}
)

func (d *DockerNetwork) addContainerCharts(name string) {
	charts := containerNetworkChartsTmpl.Copy()
	for _, chart := range *charts {
		chart.ID = fmt.Sprintf(chart.ID, name)
		chart.Labels = []module.Label{
			{Key: "container_name", Value: name},
		}
		for _, dim := range chart.Dims {
			dim.ID = fmt.Sprintf(dim.ID, name)
		}
	}

	if err := d.Charts().Add(*charts...); err != nil {
		d.Warning(err)
	}
}

func (d *DockerNetwork) removeContainerCharts(name string) {
	px := fmt.Sprintf("network_%s", name)

	for _, chart := range *d.Charts() {
		if strings.HasPrefix(chart.ID, px) {
			chart.MarkRemove()
			chart.MarkNotCreated()
		}
	}
}
