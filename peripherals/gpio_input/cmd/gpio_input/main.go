// GPIO input example: wait until an input pin reads high.
//
// Set inputPinName below to a pin of your board that is configured as an input.
package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	orbitos "github.com/OrbitOS-org/orbit-os-sdk-go/v26/client"
	"github.com/OrbitOS-org/orbit-os-sdk-go/v26/logger"
	"github.com/OrbitOS-org/orbit-os-sdk-go/v26/metadata"
)

const logTag = "gpio_input"

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
	// inputPinName must match a name returned by ListPins.
	inputPinName = "GPIO24"
	// pollInterval is how often the pin level is read.
	pollInterval = 50 * time.Millisecond
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

	// Stop cleanly on Ctrl+C, and on SIGTERM, which Orbit OS sends to stop an app.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	err = run(ctx, client)
	stop()
	client.Close()
	if err != nil {
		logger.Fatalf(logTag, "%v", err)
		os.Exit(1)
	}
	logger.Infof(logTag, "Done.")
}

func run(ctx context.Context, client *orbitos.Client) error {
	gpio := client.GpioManager

	pins, err := gpio.ListPins()
	if err != nil {
		return fmt.Errorf("ListPins: %w", err)
	}

	var pin *orbitos.GpioPin
	for _, p := range pins {
		if p.Name == inputPinName {
			pin = p
			break
		}
	}
	if pin == nil {
		return fmt.Errorf("pin %s not found on this device", inputPinName)
	}

	// This example only reads. It does not change the direction of the pin.
	dir, err := gpio.GetDirection(pin)
	if err != nil {
		return fmt.Errorf("GetDirection: %w", err)
	}
	if dir != orbitos.GPIO_DIR_IN {
		return fmt.Errorf("pin %s is not an input (direction=%v)", inputPinName, dir)
	}

	logger.Infof(logTag, "Waiting for %s to go high (Ctrl+C to stop)", inputPinName)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		level, err := gpio.GetLevel(pin)
		if err != nil {
			return fmt.Errorf("GetLevel: %w", err)
		}
		if level == orbitos.GPIO_LEVEL_HIGH {
			logger.Infof(logTag, "%s is high.", inputPinName)
			return nil
		}

		select {
		case <-ctx.Done():
			logger.Infof(logTag, "Stopped before %s went high.", inputPinName)
			return nil
		case <-ticker.C:
		}
	}
}
