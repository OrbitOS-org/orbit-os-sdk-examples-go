// Ethernet status example: list the Ethernet interfaces and print the link
// state and IP configuration of each one.
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

const logTag = "eth_status"

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
	eth := client.EthernetManager

	ifaces, err := eth.ListEthernetInterfaces()
	if err != nil {
		return fmt.Errorf("ListEthernetInterfaces: %w", err)
	}

	logger.Infof(logTag, "Ethernet interfaces: %d", len(ifaces))
	for i, iface := range ifaces {
		name := iface.GetInterfaceName()

		// A failure on one interface should not hide the others.
		linkUp, err := eth.IsEthernetConnected(name)
		if err != nil {
			logger.Errorf(logTag, "  [%d] IsEthernetConnected(%q): %v", i, name, err)
			continue
		}

		logger.Infof(logTag, "  [%d] %q linkUp=%v state=%v mac=%q mtu=%d dhcp=%v ipv4=%q gw=%q dns=[%s]",
			i, name, linkUp, iface.GetState(), iface.GetMacAddress(), iface.GetMtu(), iface.GetDhcpEnable(),
			iface.GetIpv4Address(), iface.GetIpv4Gateway(), strings.Join(iface.GetIpv4Dns(), ","))
	}
	return nil
}
