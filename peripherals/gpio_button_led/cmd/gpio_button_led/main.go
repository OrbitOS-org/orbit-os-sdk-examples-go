// GPIO button and LED example: the LED is on while the button is pressed.
//
// It reads an input pin and writes its level to an output pin, until Ctrl+C.
// Set the two pin names below to match your wiring.
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

const logTag = "gpio_button_led"

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
	// buttonPinName and ledPinName must match names returned by ListPins.
	buttonPinName = "GPIO24"
	ledPinName    = "GPIO23"
	// pollInterval is how often the button is read.
	pollInterval = 20 * time.Millisecond
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
	button, led := findPin(pins, buttonPinName), findPin(pins, ledPinName)
	if button == nil {
		return fmt.Errorf("pin %s not found on this device", buttonPinName)
	}
	if led == nil {
		return fmt.Errorf("pin %s not found on this device", ledPinName)
	}

	if err := gpio.SetDirection(button, orbitos.GPIO_DIR_IN); err != nil {
		return fmt.Errorf("SetDirection %s: %w", buttonPinName, err)
	}
	if err := gpio.SetDirection(led, orbitos.GPIO_DIR_OUT); err != nil {
		return fmt.Errorf("SetDirection %s: %w", ledPinName, err)
	}

	// Whatever happens next, leave the LED off.
	defer func() {
		if err := gpio.SetLevel(led, orbitos.GPIO_LEVEL_LOW); err != nil {
			logger.Errorf(logTag, "Turn the LED off on exit: %v", err)
		}
	}()

	logger.Infof(logTag, "%s (button) drives %s (LED). Ctrl+C to stop.", buttonPinName, ledPinName)

	// Write to the LED only when the button changes, not on every read.
	last := orbitos.GpioLevel(-1)
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		level, err := gpio.GetLevel(button)
		if err != nil {
			return fmt.Errorf("GetLevel: %w", err)
		}
		if level != last {
			if err := gpio.SetLevel(led, level); err != nil {
				return fmt.Errorf("SetLevel: %w", err)
			}
			if level == orbitos.GPIO_LEVEL_HIGH {
				logger.Infof(logTag, "Button pressed: LED on")
			} else {
				logger.Infof(logTag, "Button released: LED off")
			}
			last = level
		}

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

// findPin returns the pin with the given name, or nil.
func findPin(pins []*orbitos.GpioPin, name string) *orbitos.GpioPin {
	for _, p := range pins {
		if p.Name == name {
			return p
		}
	}
	return nil
}
