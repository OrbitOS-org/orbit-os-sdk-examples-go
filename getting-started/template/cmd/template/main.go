// Template: the smallest useful Orbit OS app. It connects to the device and
// reads one value.
//
// To start your own app: copy this project folder, rename it and cmd/template,
// update go.mod, orbit.project.json and metadata.json, then replace the body
// of run with your own code.
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

const logTag = "template"

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

// run holds the app logic. Return an error to stop the app with exit code 1.
func run(client *orbitos.Client) error {
	model, err := client.SystemManager.GetHardwareModel()
	if err != nil {
		return fmt.Errorf("GetHardwareModel: %w", err)
	}
	logger.Infof(logTag, "Hardware model: %s", model)

	// Add your code here. To call another service, add it to "permissions"
	// in metadata.json first: Orbit OS refuses calls to services that are
	// not listed there.
	return nil
}
