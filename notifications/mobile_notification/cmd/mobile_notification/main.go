// Mobile notification example: send a push notification to the phones linked
// to the device owner's Orbit OS account, with MobileNotificationManager.Send
// (added in API 26.1).
//
// The app only calls Send. Orbit OS delivers the notification through the
// Orbit OS Cloud Services; the app holds no cloud credentials.
//
// What you get back depends on the device at the time of the run:
//   - the device is not signed in to an Orbit OS account, or no phone is
//     linked to that account: an error with code ERROR_CODE_UNAVAILABLE;
//   - signed in, with at least one phone linked: sent is 1 or more.
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

const logTag = "mobile_notification"

// defaultHost is only used in development, when the app runs on your computer
// and talks to a device in Developer Mode. On the device the SDK connects locally.
const defaultHost = "192.168.1.100"

// metadata.json is the app manifest. Orbit Studio reads the same file when it
// builds the .orb package.
//
//go:embed metadata.json
var metadataJSON []byte

var appManifest = metadata.MustParseAppManifestJSON(metadataJSON)

// notificationBody is the text shown in the notification.
const notificationBody = "Device status changed"

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
	// There is no title parameter. Orbit OS always sets the title to
	// "<device name> · <this app's name>", taken from the app's identity, so
	// one app cannot send a notification that looks like it came from another.
	//
	// The second argument holds the optional fields. For example, to attach
	// data for the mobile app and replace an earlier notification:
	//
	//	&orbitos.SendOptions{
	//		Data:        map[string]string{"screen": "status"},
	//		CollapseKey: "status",
	//	}
	sent, failed, err := client.MobileNotificationManager.Send(notificationBody, nil)
	if err != nil {
		// The error message carries the code returned by Orbit OS,
		// for example "... (code: ERROR_CODE_UNAVAILABLE)".
		return fmt.Errorf("Send: %w", err)
	}

	// sent and failed count the phones the notification was sent to.
	logger.Infof(logTag, "Notification sent: sent=%d failed=%d", sent, failed)
	return nil
}
