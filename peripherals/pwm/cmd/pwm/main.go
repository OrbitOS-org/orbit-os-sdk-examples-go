// PWM example: list the PWM channels, set the first one to a fixed frequency
// and duty cycle, then read the settings back.
//
// The channel keeps running with these settings after the example ends.
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

const logTag = "pwm"

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
	// dutyCycle is the fraction of each period the output stays high (0.5 = 50%).
	dutyCycle = 0.5
	// frequencyHz is the output frequency.
	frequencyHz = 10.0
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
	pwm := client.PwmManager

	channels, err := pwm.ListChannels()
	if err != nil {
		return fmt.Errorf("ListChannels: %w", err)
	}

	logger.Infof(logTag, "PWM channels: %d", len(channels))
	for i, ch := range channels {
		logger.Infof(logTag, "  [%d] name=%q channel=%d", i, ch.Name, ch.Channel)
	}
	if len(channels) == 0 {
		logger.Infof(logTag, "No PWM channel on this device: nothing to configure.")
		return nil
	}

	channel := channels[0]
	logger.Infof(logTag, "Setting %q to duty=%.0f%% frequency=%g Hz", channel.Name, dutyCycle*100, frequencyHz)
	if err := pwm.SetPwm(channel, dutyCycle, frequencyHz); err != nil {
		return fmt.Errorf("SetPwm: %w", err)
	}

	// Read the settings back to confirm what the device applied.
	props, err := pwm.GetProperties(channel)
	if err != nil {
		return fmt.Errorf("GetProperties: %w", err)
	}
	logger.Infof(logTag, "Read back: enabled=%v duty=%.3f frequency=%.3f Hz", props.Enabled, props.DutyCycle, props.FrequencyHz)
	return nil
}
