// UART example: open a serial port at 9600 8N1 and print what it receives,
// until Ctrl+C.
//
// The received bytes go to stdout as they arrive; the log lines go through the
// SDK logger. In a terminal the two may interleave.
//
// Set the constants below to match your device.
package main

import (
	"context"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	orbitos "github.com/OrbitOS-org/orbit-os-sdk-go/v26/client"
	"github.com/OrbitOS-org/orbit-os-sdk-go/v26/logger"
	"github.com/OrbitOS-org/orbit-os-sdk-go/v26/metadata"
)

const logTag = "uart"

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
	// uartPortName must match a name returned by ListPorts.
	uartPortName = "ttyUSB0"
	// uartBaudRate is the line speed.
	uartBaudRate = 9600
	// rxChunkSize is the most bytes handed to the receive callback at once.
	rxChunkSize = 256
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
	uart := client.UartManager

	ports, err := uart.ListPorts()
	if err != nil {
		return fmt.Errorf("ListPorts: %w", err)
	}
	logger.Infof(logTag, "UART ports: %v", ports)

	found := false
	for _, p := range ports {
		if p == uartPortName {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("port %s not found on this device (available: %v)", uartPortName, ports)
	}

	logger.Infof(logTag, "Opening %s at %d 8N1 (Ctrl+C to stop)", uartPortName, uartBaudRate)
	port, err := uart.Open(orbitos.UartConfig{
		Port:        uartPortName,
		Baudrate:    uartBaudRate,
		DataBits:    8,
		Parity:      orbitos.UartParityNone,
		StopBits:    orbitos.UartStopBits1,
		FlowControl: orbitos.UartFlowNone,
	})
	if err != nil {
		return fmt.Errorf("Open %s: %w", uartPortName, err)
	}

	// Release the port when done: only one app can hold a UART port at a time.
	defer func() {
		if err := port.Close(); err != nil {
			logger.Errorf(logTag, "Close: %v", err)
		}
	}()

	// Read the configuration back to confirm what the device applied.
	cfg, err := port.GetConfig()
	if err != nil {
		return fmt.Errorf("GetConfig: %w", err)
	}
	logger.Infof(logTag, "Active config: port=%s baud=%d dataBits=%d parity=%d stopBits=%d flow=%d",
		cfg.Port, cfg.Baudrate, cfg.DataBits, cfg.Parity, cfg.StopBits, cfg.FlowControl)

	// Listen blocks until ctx is cancelled, and calls onReceive for each chunk.
	err = port.Listen(ctx, rxChunkSize, onReceive)
	if err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("Listen: %w", err)
	}
	return nil
}

// onReceive writes the received bytes to stdout as they are.
func onReceive(chunk []byte) {
	if _, err := os.Stdout.Write(chunk); err != nil {
		logger.Errorf(logTag, "Write to stdout: %v", err)
	}
}
