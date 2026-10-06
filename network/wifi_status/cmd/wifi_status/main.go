// WiFi status example: for each WiFi interface, show whether it is connected,
// its mode, and the details of the connection (client mode) or of the access
// point (AP mode).
//
// The example only reads. It does not change the WiFi configuration.
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

const logTag = "wifi_status"

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
	wifi := client.WiFiManager

	ifaces, err := wifi.ListInterfaces()
	if err != nil {
		return fmt.Errorf("ListInterfaces: %w", err)
	}
	if len(ifaces) == 0 {
		logger.Infof(logTag, "No WiFi interface on this device.")
		return nil
	}

	for _, iface := range ifaces {
		name := iface.GetInterfaceName()
		logger.Infof(logTag, "─── %s", name)

		connected, err := wifi.IsConnected(name)
		if err != nil {
			return fmt.Errorf("IsConnected(%s): %w", name, err)
		}
		mode, err := wifi.GetMode(name)
		if err != nil {
			return fmt.Errorf("GetMode(%s): %w", name, err)
		}
		logger.Infof(logTag, "connected=%v mode=%v state=%v mac=%q", connected, mode, iface.GetState(), iface.GetMacAddress())

		// The interface carries the details of the mode it is in: client
		// properties when it joined a network, AP properties when it is an
		// access point. The other one is nil.
		if c := iface.GetClientProperties(); c != nil {
			logger.Infof(logTag, "client: ssid=%q bssid=%q security=%q rssi=%d ch=%d band=%q",
				c.GetSsid(), c.GetBssid(), c.GetSecurity(), c.GetSignalStrength(), c.GetChannel(), c.GetBand())
			logger.Infof(logTag, "client: dhcp=%v ipv4=%q gw=%q dns=[%s]",
				c.GetDhcpEnable(), c.GetIpv4Address(), c.GetIpv4Gateway(), strings.Join(c.GetIpv4Dns(), ","))
		}
		if ap := iface.GetApProperties(); ap != nil {
			logger.Infof(logTag, "access point: ssid=%q security=%q ipv4=%q ch=%d band=%q hidden=%v active=%v clients=%d",
				ap.GetSsid(), ap.GetSecurity(), ap.GetIpv4Address(), ap.GetChannel(), ap.GetBand(),
				ap.GetHidden(), ap.GetActive(), ap.GetConnectedClients())
		}
	}
	return nil
}
