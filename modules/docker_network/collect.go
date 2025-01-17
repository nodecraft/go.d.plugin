package docker_network

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"strings"
	"time"
)

func (d *DockerNetwork) collect() (map[string]int64, error) {
	if d.client == nil {
		// Create a new client
		client, err := d.newClient(d.Config)
		if err != nil {
			return nil, err
		}
		d.client = client
	}

	// Make sure we've negotiated the API version
	if !d.verNegotiated {
		d.verNegotiated = true
		d.negotiateAPIVersion()
	}

	// Defer closing the client
	defer func() { _ = d.client.Close() }()

	mx := make(map[string]int64)

	// Collect our info
	if err := d.collectContainers(mx); err != nil {
		return nil, err
	}

	return mx, nil
}

func (d *DockerNetwork) collectContainers(mx map[string]int64) error {
	// This function will collect all the containers network stats

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout.Duration)
	defer cancel()

	// Get all the containers
	containers, err := d.client.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return err
	}

	seen := make(map[string]bool)

	// We need to do the stats retrieval async
	// We need some way to know when all of them are done
	// So we use a waitgroup

	waitGroup := make(chan struct{}, len(containers))

	// Loop through all the containers and get their stats
	for _, container := range containers {
		// Add to the waitgroup
		waitGroup <- struct{}{}
		// Create a new context for each container
		ctx, cancel := context.WithTimeout(context.Background(), d.Timeout.Duration+(time.Second*5))
		go func(ctx context.Context, container types.Container) {
			d.Debugf("collecting stats for container %s", container.ID[:12])
			defer func() {
				// We're done with this container
				<-waitGroup
				// Close the context
				cancel()
			}()
			stats, err := d.client.ContainerStats(ctx, container.ID, false)
			if err != nil {
				return
			}

			// This returns a body that's a reader, so we need to read it
			body := stats.Body

			// Now we can decode the stats
			var stat types.StatsJSON
			if err := json.NewDecoder(body).Decode(&stat); err != nil {
				return
			}
			// Close the body
			if err := body.Close(); err != nil {
				return
			}
			// We can now get the network stats
			network := stat.Networks
			// If there's no previous stats for this container then we ignore it but save the stats
			if _, ok := d.previousStats[container.ID]; !ok {
				d.previousStats[container.ID] = stat
				return
			}
			// Now we want the tx and rx bytes
			txBytes := 0
			rxBytes := 0
			// Loop through the networks and add em up
			for _, net := range network {
				txBytes += int(net.TxBytes)
				rxBytes += int(net.RxBytes)
			}
			// Add up previous stats
			prev := d.previousStats[container.ID]
			for _, net := range prev.Networks {
				txBytes -= int(net.TxBytes)
				rxBytes -= int(net.RxBytes)
			}
			// Now we have how much traffic has happened in the last "update time" seconds
			// So to get this into a bytes/sec we divide by the update time
			txBytes /= d.UpdateEvery
			rxBytes /= d.UpdateEvery
			// Now we have what we wanted
			name := strings.TrimPrefix(container.Names[0], "/")

			seen[name] = true

			if !d.containers[name] {
				// Add the container to our charts
				d.addContainerCharts(name)
				d.containers[name] = true
			}

			// Now we create our metrics
			px := fmt.Sprintf("container_%s_", name)
			mx[px+"network_bytes_tx"] = int64(txBytes)
			mx[px+"network_bytes_rx"] = int64(rxBytes)
			// Update the previous stats
			d.previousStats[container.ID] = stat
			d.Debugf("collected stats for container %s", container.ID[:12])
		}(ctx, container)
	}

	// Wait for all the containers to be done
	for i := 0; i < cap(waitGroup); i++ {
		waitGroup <- struct{}{}
	}

	for name := range d.containers {
		if !seen[name] {
			delete(d.containers, name)
			d.removeContainerCharts(name)
		}
	}

	return nil
}

func (d *DockerNetwork) negotiateAPIVersion() {
	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout.Duration)
	defer cancel()

	d.client.NegotiateAPIVersion(ctx)
}
