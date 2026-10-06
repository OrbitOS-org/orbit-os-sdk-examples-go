// Camera snapshot example: list the cameras, lock the first one, take one
// picture, save it to a file and release the camera.
//
// On the device the file is written to the app's data folder, which is the
// working directory of the app.
package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	orbitos "github.com/OrbitOS-org/orbit-os-sdk-go/v26/client"
	"github.com/OrbitOS-org/orbit-os-sdk-go/v26/logger"
	"github.com/OrbitOS-org/orbit-os-sdk-go/v26/metadata"
)

const logTag = "camera_snapshot"

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
	// captureWidth, captureHeight and captureFormat describe the picture to
	// take. The camera must support them: see the formats and resolutions the
	// example prints for each camera.
	captureWidth  = 640
	captureHeight = 480
	captureFormat = "mjpeg"
	// outputFile is where the picture is saved. An MJPEG frame is a JPEG image.
	outputFile = "snapshot.jpg"
	// timeout is the longest the whole capture may take.
	timeout = 15 * time.Second
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
	camera := client.CameraManager

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	devices, err := camera.ListDevices(ctx)
	if err != nil {
		return fmt.Errorf("ListDevices: %w", err)
	}

	logger.Infof(logTag, "Cameras: %d", len(devices))
	for i, d := range devices {
		logger.Infof(logTag, "  [%d] id=%q card=%q driver=%q formats=%v resolutions=%v",
			i, d.DeviceID, d.Card, d.Driver, d.SupportedFormats, d.Resolutions)
	}
	if len(devices) == 0 {
		logger.Infof(logTag, "No camera on this device: nothing to capture.")
		return nil
	}

	// Only one app can hold a camera at a time. Lock it, use it, unlock it.
	deviceID := devices[0].DeviceID
	locked, err := camera.Lock(ctx, deviceID, logTag)
	if err != nil {
		return fmt.Errorf("Lock %s: %w", deviceID, err)
	}
	defer func() {
		// Unlock with its own context, so it still runs after a timeout.
		unlockCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := locked.Unlock(unlockCtx); err != nil {
			logger.Errorf(logTag, "Unlock: %v", err)
		}
	}()

	logger.Infof(logTag, "Capturing %dx%d %s from %s", captureWidth, captureHeight, captureFormat, deviceID)
	image, err := locked.CaptureImage(ctx, captureWidth, captureHeight, captureFormat)
	if err != nil {
		return fmt.Errorf("CaptureImage: %w", err)
	}
	if len(image.ImageData) == 0 {
		return fmt.Errorf("the camera returned an empty image")
	}

	if err := os.WriteFile(outputFile, image.ImageData, 0o644); err != nil {
		return fmt.Errorf("save %s: %w", outputFile, err)
	}
	logger.Infof(logTag, "Saved %s (%d bytes, format %s)", outputFile, len(image.ImageData), image.Format)
	return nil
}
