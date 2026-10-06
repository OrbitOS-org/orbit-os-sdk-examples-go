// Basic example: connect to the device and read system information —
// API version, device identity, hardware, software versions and live metrics.
//
// Every value comes from SystemManager. The table in run lists each getter by
// name, so the source doubles as a map of what the system service offers.
package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"strings"

	orbitos "github.com/OrbitOS-org/orbit-os-sdk-go/v26/client"
	"github.com/OrbitOS-org/orbit-os-sdk-go/v26/logger"
	"github.com/OrbitOS-org/orbit-os-sdk-go/v26/metadata"
)

const logTag = "basic"

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
	sys := client.SystemManager

	// The API version belongs to the device. It is not the version of the SDK
	// this app was built with.
	apiVer, apiRev, err := sys.GetApiVersion()
	if err != nil {
		return fmt.Errorf("GetApiVersion: %w", err)
	}
	logger.Infof(logTag, "Device API version: %d.%d", apiVer, apiRev)

	type field struct {
		name string
		get  func() (string, error)
	}
	sections := []struct {
		title  string
		fields []field
	}{
		{"Device identity", []field{
			{"DeviceName", sys.GetDeviceName},
			{"SocModel", sys.GetSocModel},
			{"SocVendor", sys.GetSocVendor},
			{"BoardModel", sys.GetBoardModel},
			{"BoardVendor", sys.GetBoardVendor},
			{"HardwareModel", sys.GetHardwareModel},
			{"HardwareVersion", sys.GetHardwareVersion},
			{"SystemUuid", sys.GetSystemUuid},
			{"BoardSerial", sys.GetBoardSerial},
			{"CpuSerial", sys.GetCpuSerial},
			{"MachineId", sys.GetMachineId},
			{"Architecture", sys.GetArchitecture},
			{"CpuModel", sys.GetCpuModel},
		}},
		{"Software", []field{
			{"OsName", sys.GetOsName},
			{"OsVersion", sys.GetOsVersion},
			{"OsRevision", sys.GetOsRevision},
			{"KernelVersion", sys.GetKernelVersion},
			{"Distro", sys.GetDistro},
			{"DistroVersion", sys.GetDistroVersion},
			{"RuntimeVersion", sys.GetRuntimeVersion},
			{"RuntimeBuildDate", sys.GetRuntimeBuildDate},
		}},
	}
	for _, s := range sections {
		logger.Infof(logTag, "─── %s", s.title)
		for _, f := range s.fields {
			v, err := f.get()
			if err != nil {
				return fmt.Errorf("Get%s: %w", f.name, err)
			}
			logger.Infof(logTag, "%-17s %s", f.name+":", v)
		}
	}

	logger.Infof(logTag, "─── CPU and memory")

	ram, err := sys.GetTotalRAM()
	if err != nil {
		return fmt.Errorf("GetTotalRAM: %w", err)
	}
	logger.Infof(logTag, "%-17s %d", "TotalRAM:", ram)

	cores, err := sys.GetCpuCores()
	if err != nil {
		return fmt.Errorf("GetCpuCores: %w", err)
	}
	logger.Infof(logTag, "%-17s %d", "CpuCores:", cores)

	threads, err := sys.GetCpuThreads()
	if err != nil {
		return fmt.Errorf("GetCpuThreads: %w", err)
	}
	logger.Infof(logTag, "%-17s %d", "CpuThreads:", threads)

	minMhz, err := sys.GetCpuMinMhz()
	if err != nil {
		return fmt.Errorf("GetCpuMinMhz: %w", err)
	}
	logger.Infof(logTag, "%-17s %.1f", "CpuMinMhz:", minMhz)

	maxMhz, err := sys.GetCpuMaxMhz()
	if err != nil {
		return fmt.Errorf("GetCpuMaxMhz: %w", err)
	}
	logger.Infof(logTag, "%-17s %.1f", "CpuMaxMhz:", maxMhz)

	// One snapshot of the live metrics. The metrics_info example shows more.
	logger.Infof(logTag, "─── Metrics")

	resp, err := sys.GetMetrics()
	if err != nil {
		return fmt.Errorf("GetMetrics: %w", err)
	}
	m := resp.GetMetrics()
	if m == nil {
		logger.Infof(logTag, "The device returned no metrics.")
		return nil
	}
	logger.Infof(logTag, "%-17s %.1f%%", "CpuUsage:", m.GetCpuUsage())
	logger.Infof(logTag, "%-17s %.1f%%", "MemoryUsage:", m.GetMemoryUsage())
	logger.Infof(logTag, "%-17s %.1f°C", "SocThermal:", m.GetSocThermal())
	logger.Infof(logTag, "%-17s %d°C", "SensorsTemp:", m.GetSensorsTemperatures())
	if net := m.GetNetwork(); net != nil {
		logger.Infof(logTag, "%-17s %d", "Network RxBytes:", net.GetBytesRecv())
		logger.Infof(logTag, "%-17s %d", "Network TxBytes:", net.GetBytesSent())
	}
	return nil
}
