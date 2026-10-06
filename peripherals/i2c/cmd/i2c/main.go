// I2C example: list the I2C buses, open the first one, read its configuration
// and scan it for devices.
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

const logTag = "i2c"

// defaultHost is only used in development, when the app runs on your computer
// and talks to a device in Developer Mode. On the device the SDK connects locally.
const defaultHost = "192.168.1.100"

// metadata.json is the app manifest. Orbit Studio reads the same file when it
// builds the .orb package.
//
//go:embed metadata.json
var metadataJSON []byte

var appManifest = metadata.MustParseAppManifestJSON(metadataJSON)

const (
	// clockHz is the bus speed: 100000 is standard mode (100 kHz).
	clockHz = 100000
	// tenBitAddr selects 10-bit addressing. Most devices use 7-bit addresses.
	tenBitAddr = false
	// clockStretching lets a slow device hold the clock line while it gets ready.
	clockStretching = true
)

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
	i2c := client.I2CManager

	buses, err := i2c.ListBuses()
	if err != nil {
		return fmt.Errorf("ListBuses: %w", err)
	}

	logger.Infof(logTag, "I2C buses: %v", buses)
	if len(buses) == 0 {
		logger.Infof(logTag, "No I2C bus on this device: nothing to do.")
		return nil
	}

	busNumber := buses[0]
	logger.Infof(logTag, "Using bus %d", busNumber)

	bus, err := i2c.Open(busNumber, clockHz, tenBitAddr, clockStretching)
	if err != nil {
		return fmt.Errorf("Open bus %d: %w", busNumber, err)
	}

	cfg, err := bus.GetConfig()
	if err != nil {
		return fmt.Errorf("GetConfig: %w", err)
	}
	logger.Infof(logTag, "Config: clockHz=%d tenBit=%v clockStretching=%v", cfg.ClockHz, cfg.TenBitAddr, cfg.ClockStretching)

	addrs, err := bus.Scan()
	if err != nil {
		return fmt.Errorf("Scan: %w", err)
	}
	if len(addrs) == 0 {
		logger.Infof(logTag, "No device answered on bus %d.", busNumber)
		return nil
	}
	logger.Infof(logTag, "%d device(s) on bus %d:", len(addrs), busNumber)
	for _, addr := range addrs {
		logger.Infof(logTag, "  address 0x%02X", addr)
	}

	// To read from a device, use one of the addresses found above. For example,
	// to read 1 byte from the device at 0x48:
	//
	//	data, err := bus.Transfer(0x48, nil, 1, 0)
	return nil
}
