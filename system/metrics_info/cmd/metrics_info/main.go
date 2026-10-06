// System metrics example: read one snapshot of the device metrics — CPU,
// memory, disk and temperature — with SystemManager.GetMetrics.
package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	orbitos "github.com/OrbitOS-org/orbit-os-sdk-go/v26/client"
	"github.com/OrbitOS-org/orbit-os-sdk-go/v26/logger"
	"github.com/OrbitOS-org/orbit-os-sdk-go/v26/metadata"
)

const logTag = "metrics_info"

// defaultHost is only used in development, when the app runs on your computer
// and talks to a device in Developer Mode. On the device the SDK connects locally.
const defaultHost = "192.168.1.100"

// metadata.json is the app manifest. Orbit Studio reads the same file when it
// builds the .orb package.
//
//go:embed metadata.json
var metadataJSON []byte

var appManifest = metadata.MustParseAppManifestJSON(metadataJSON)

func main() {
	host := flag.String("host", defaultHost, "Device IP address (development only; ignored when the app runs on the device)")
	flag.Parse()

	meta := metadata.Build(appManifest)
	logger.Init(meta.Name, "INFO", true)
	logger.Infof(logTag, "Starting %s v%s", meta.Name, meta.Version)
	appManifest.PrintInfo()

	client, err := orbitos.NewClientAuto(strings.TrimSpace(*host))
	if err != nil {
		logger.Fatalf(logTag, "Connect: %v", err)
		os.Exit(1)
	}

	err = run(client)
	client.Close()
	if err != nil {
		logger.Fatalf(logTag, "%v", err)
		os.Exit(1)
	}
	logger.Infof(logTag, "Done.")
}

func run(client *orbitos.Client) error {
	resp, err := client.SystemManager.GetMetrics()
	if err != nil {
		return fmt.Errorf("GetMetrics: %w", err)
	}
	m := resp.GetMetrics()
	if m == nil {
		logger.Infof(logTag, "The device returned no metrics.")
		return nil
	}

	// CPU: overall usage, then one value per core.
	logger.Infof(logTag, "cpu: usage=%.1f%% cores=%s", m.GetCpuUsage(), formatPercents(m.GetCpuCorePercent()))

	logger.Infof(logTag, "memory: usage=%.1f%%", m.GetMemoryUsage())

	// Disk usage comes as a map of mount point to usage. Sort the mount
	// points so the output is the same on every run.
	disks := m.GetDiskUsage()
	paths := make([]string, 0, len(disks))
	for path := range disks {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		logger.Infof(logTag, "disk %s: usage=%.1f%%", path, disks[path].GetPercent())
	}

	logger.Infof(logTag, "thermal: soc=%.1f°C sensors=%d°C", m.GetSocThermal(), m.GetSensorsTemperatures())
	return nil
}

// formatPercents formats a list of percentages as "[12.50 3.10]".
func formatPercents(values []float64) string {
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = fmt.Sprintf("%.2f", v)
	}
	return "[" + strings.Join(parts, " ") + "]"
}
