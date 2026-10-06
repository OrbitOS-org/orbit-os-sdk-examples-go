// PWM fade example: fade the first PWM channel in and out, like a breathing
// LED, until Ctrl+C. The output is stopped when the example ends.
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

const logTag = "pwm_fade"

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
	// frequencyHz is the output frequency. 1 kHz is fast enough for an LED
	// not to flicker.
	frequencyHz = 1000.0
	// steps is the number of duty cycle levels between off and fully on.
	steps = 50
	// stepInterval is the time spent on each level.
	stepInterval = 20 * time.Millisecond
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
	pwm := client.PwmManager

	channels, err := pwm.ListChannels()
	if err != nil {
		return fmt.Errorf("ListChannels: %w", err)
	}
	if len(channels) == 0 {
		logger.Infof(logTag, "No PWM channel on this device: nothing to do.")
		return nil
	}

	channel := channels[0]
	logger.Infof(logTag, "Fading %q at %g Hz. Ctrl+C to stop.", channel.Name, frequencyHz)

	// Stop the output when done, so the channel is not left running.
	defer func() {
		if err := pwm.StopPwm(channel); err != nil {
			logger.Errorf(logTag, "StopPwm: %v", err)
		}
	}()

	ticker := time.NewTicker(stepInterval)
	defer ticker.Stop()

	// step goes 0, 1, … steps, then back down to 0, and repeats.
	step, direction := 0, 1
	for {
		duty := float64(step) / steps
		if err := pwm.SetPwm(channel, duty, frequencyHz); err != nil {
			return fmt.Errorf("SetPwm: %w", err)
		}

		if step == steps {
			direction = -1
		} else if step == 0 {
			direction = 1
		}
		step += direction

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}
