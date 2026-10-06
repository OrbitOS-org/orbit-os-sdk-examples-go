// Notification options example: send a push notification with the optional
// fields of MobileNotificationManager.Send (added in API 26.1).
//
// See the mobile_notification example first: it sends the simplest
// notification and explains what the result means.
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

const logTag = "notification_options"

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
	// notificationBody is the text shown in the notification.
	notificationBody = "Water tank level is low"
	// collapseKey groups notifications: a new one replaces an earlier one with
	// the same key that the phone has not shown yet.
	collapseKey = "tank-level"
	// ttlSeconds is how long the notification is kept for a phone that is
	// offline. After that it is dropped.
	ttlSeconds = 3600
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
	options := &orbitos.SendOptions{
		// Data is delivered to the mobile app with the notification, for
		// example to open a specific screen. It is not shown to the user.
		Data: map[string]string{
			"screen": "tank",
			"level":  "12",
		},
		CollapseKey: collapseKey,
		TtlSeconds:  ttlSeconds,
	}

	sent, failed, err := client.MobileNotificationManager.Send(notificationBody, options)
	if err != nil {
		return fmt.Errorf("Send: %w", err)
	}

	// sent and failed count the phones the notification was sent to.
	logger.Infof(logTag, "Notification sent: sent=%d failed=%d", sent, failed)
	return nil
}
