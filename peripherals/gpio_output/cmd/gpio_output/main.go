// GPIO output example: list the GPIO lines of the device, then blink the
// first one.
//
// Check which line comes first on your board before you wire anything to it.
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

const logTag = "gpio_output"

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
	// blinkPhase is how long the line stays high, and then low, in each cycle.
	blinkPhase = 500 * time.Millisecond
	// blinkCycles is the number of high-then-low cycles.
	blinkCycles = 20
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

	logger.Infof(logTag, "GPIO lines: %d", len(pins))
	for i, p := range pins {
		logger.Infof(logTag, "  [%d] name=%q gpiochip=%d line=%d", i, p.Name, p.ChipNumber, p.Number)
	}
	if len(pins) == 0 {
		logger.Infof(logTag, "No GPIO line on this device: nothing to blink.")
		return nil
	}

	pin := pins[0]
	logger.Infof(logTag, "Blinking %q (gpiochip=%d line=%d)", pin.Name, pin.ChipNumber, pin.Number)

	if err := gpio.SetDirection(pin, orbitos.GPIO_DIR_OUT); err != nil {
		return fmt.Errorf("SetDirection: %w", err)
	}

	// Whatever happens next, leave the line low.
	defer func() {
		if err := gpio.SetLevel(pin, orbitos.GPIO_LEVEL_LOW); err != nil {
			logger.Errorf(logTag, "Set the line low on exit: %v", err)
		}
	}()

	for cycle := 1; cycle <= blinkCycles; cycle++ {
		if err := gpio.SetLevel(pin, orbitos.GPIO_LEVEL_HIGH); err != nil {
			return fmt.Errorf("SetLevel high: %w", err)
		}
		logger.Infof(logTag, "  cycle %d/%d high", cycle, blinkCycles)
		if !sleep(ctx, blinkPhase) {
			return nil
		}

		if err := gpio.SetLevel(pin, orbitos.GPIO_LEVEL_LOW); err != nil {
			return fmt.Errorf("SetLevel low: %w", err)
		}
		logger.Infof(logTag, "  cycle %d/%d low", cycle, blinkCycles)
		if !sleep(ctx, blinkPhase) {
			return nil
		}
	}
	return nil
}

// sleep waits for d. It returns false if the app was asked to stop first.
func sleep(ctx context.Context, d time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}
