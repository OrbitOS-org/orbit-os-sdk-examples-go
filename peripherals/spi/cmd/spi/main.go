// SPI example: list the SPI devices, open one, read its configuration and run
// a full-duplex transfer.
//
// Set the constants below to match your wiring.
package main

import (
	"bytes"
	_ "embed"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strings"

	orbitos "github.com/OrbitOS-org/orbit-os-sdk-go/v26/client"
	"github.com/OrbitOS-org/orbit-os-sdk-go/v26/logger"
	"github.com/OrbitOS-org/orbit-os-sdk-go/v26/metadata"
)

const logTag = "spi"

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
	// spiBus and spiCS select the device: bus 0, chip select 2 is "spidev0.2".
	spiBus uint32 = 0
	spiCS  uint32 = 2
	// spiSpeedHz is the maximum clock speed.
	spiSpeedHz uint32 = 10_000
	// spiBitsPerWord, spiMode and spiLSBFirst set the frame format.
	spiBitsPerWord uint32 = 8
	spiMode               = 0
	spiLSBFirst           = false
	// spiLoopback: with MOSI wired to MISO, set this to true to send a short
	// pattern and check that the same bytes come back.
	spiLoopback = false
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
	spi := client.SpiManager

	devices, err := spi.ListDevices()
	if err != nil {
		return fmt.Errorf("ListDevices: %w", err)
	}
	logger.Infof(logTag, "SPI devices: %v", devices)

	dev, err := spi.Open(spiBus, spiCS, spiSpeedHz, spiBitsPerWord, spiMode, spiLSBFirst)
	if err != nil {
		return fmt.Errorf("Open spidev%d.%d: %w", spiBus, spiCS, err)
	}

	cfg, err := dev.GetConfig()
	if err != nil {
		return fmt.Errorf("GetConfig: %w", err)
	}
	logger.Infof(logTag, "Config spidev%d.%d: speed=%d bits=%d mode=%d lsbFirst=%v",
		spiBus, spiCS, cfg.MaxSpeedHz, cfg.BitsPerWord, cfg.Mode, cfg.LSBFirst)

	// SPI is full duplex: for every byte sent, one byte is received.
	tx := []byte{0xD1, 0x00, 0x00, 0x01, 0xFF, 0xFF}
	if spiLoopback {
		tx = []byte{0xAB, 0xCD, 0x12, 0x34}
	}

	rx, err := dev.Transfer(tx, uint32(len(tx)))
	if err != nil {
		return fmt.Errorf("Transfer: %w", err)
	}
	logger.Infof(logTag, "TX %s", hex.EncodeToString(tx))
	logger.Infof(logTag, "RX %s", hex.EncodeToString(rx))

	if spiLoopback && !bytes.Equal(tx, rx) {
		return fmt.Errorf("loopback failed: sent %x, received %x", tx, rx)
	}
	return nil
}
