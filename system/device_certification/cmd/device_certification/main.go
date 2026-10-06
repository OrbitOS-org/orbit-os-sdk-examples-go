// Device certification example: check whether the app runs on certified
// hardware, with IsDeviceCertified and GetSerialNumber (added in API 26.1).
//
// Expected output on a certified device:
//
//	IsDeviceCertified: true
//	SerialNumber:      <certification serial>
//
// Expected output on a community board (for example a Raspberry Pi):
//
//	IsDeviceCertified: false
//	SerialNumber:      (empty)
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

const logTag = "device_certification"

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

	// These calls need a device with API 26.1 or newer.
	apiVer, apiRev, err := sys.GetApiVersion()
	if err != nil {
		return fmt.Errorf("GetApiVersion: %w", err)
	}
	logger.Infof(logTag, "Device API version: %d.%d", apiVer, apiRev)

	logger.Infof(logTag, "─── Device certification")

	certified, err := sys.IsDeviceCertified()
	if err != nil {
		return fmt.Errorf("IsDeviceCertified: %w", err)
	}
	logger.Infof(logTag, "IsDeviceCertified: %v", certified)

	// The certification serial is empty on a device that is not certified.
	serial, err := sys.GetSerialNumber()
	if err != nil {
		return fmt.Errorf("GetSerialNumber: %w", err)
	}
	if serial == "" {
		logger.Infof(logTag, "SerialNumber:      (empty)")
	} else {
		logger.Infof(logTag, "SerialNumber:      %s", serial)
	}

	// Board information, for comparison. BoardSerial is the serial of the
	// hardware itself; it is not the certification serial.
	logger.Infof(logTag, "─── Board information")

	vendor, err := sys.GetBoardVendor()
	if err != nil {
		return fmt.Errorf("GetBoardVendor: %w", err)
	}
	logger.Infof(logTag, "BoardVendor:       %s", vendor)

	model, err := sys.GetBoardModel()
	if err != nil {
		return fmt.Errorf("GetBoardModel: %w", err)
	}
	logger.Infof(logTag, "BoardModel:        %s", model)

	boardSerial, err := sys.GetBoardSerial()
	if err != nil {
		return fmt.Errorf("GetBoardSerial: %w", err)
	}
	logger.Infof(logTag, "BoardSerial:       %s", boardSerial)

	if certified {
		logger.Infof(logTag, "This is a certified device (serial %s).", serial)
	} else {
		logger.Infof(logTag, "This is a community board: it has no certification serial.")
	}
	return nil
}
