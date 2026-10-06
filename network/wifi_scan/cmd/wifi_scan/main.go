// WiFi scan example: list the WiFi interfaces, then scan for networks on the
// first one.
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

const logTag = "wifi_scan"

// defaultHost is only used in development, when the app runs on your computer
// and talks to a device in Developer Mode. On the device the SDK connects locally.
const defaultHost = "192.168.1.100"

// metadata.json is the app manifest. Orbit Studio reads the same file when it
// builds the .orb package.
//
//go:embed metadata.json
var metadataJSON []byte

var appManifest = metadata.MustParseAppManifestJSON(metadataJSON)

// forceRescan asks the hardware for a fresh scan, which takes about 3 seconds.
// Set it to false to get the cached results at once.
const forceRescan = true

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
	wifi := client.WiFiManager

	ifaces, err := wifi.ListInterfaces()
	if err != nil {
		return fmt.Errorf("ListInterfaces: %w", err)
	}

	logger.Infof(logTag, "WiFi interfaces: %d", len(ifaces))
	for i, iface := range ifaces {
		logger.Infof(logTag, "  [%d] name=%q mac=%q mode=%v state=%v",
			i, iface.GetInterfaceName(), iface.GetMacAddress(), iface.GetMode(), iface.GetState())
	}
	if len(ifaces) == 0 {
		logger.Infof(logTag, "No WiFi interface on this device: nothing to scan.")
		return nil
	}

	ifname := ifaces[0].GetInterfaceName()
	logger.Infof(logTag, "Scanning on %q (forceRescan=%v)", ifname, forceRescan)

	networks, err := wifi.Scan(ifname, forceRescan)
	if err != nil {
		return fmt.Errorf("Scan: %w", err)
	}

	logger.Infof(logTag, "Networks found: %d", len(networks))
	for i, n := range networks {
		logger.Infof(logTag, "  [%d] ssid=%q bssid=%q security=%q rssi=%d ch=%d band=%q hidden=%v",
			i, n.GetSsid(), n.GetBssid(), n.GetSecurity(), n.GetSignalStrength(),
			n.GetChannel(), n.GetBand(), n.GetHidden())
	}
	return nil
}
