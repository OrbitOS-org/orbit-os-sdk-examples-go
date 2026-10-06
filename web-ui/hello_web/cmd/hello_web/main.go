// Hello web example: an app with a web page, opened from the Orbit OS Launcher.
//
// The app listens on 127.0.0.1 only and registers the page with the Launcher.
// The Launcher serves it at http://<device>/hello_web, asks for the device
// login and forwards the requests to the app. Nothing is reachable from the
// network without going through the Launcher.
package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	orbitos "github.com/OrbitOS-org/orbit-os-sdk-go/v26/client"
	"github.com/OrbitOS-org/orbit-os-sdk-go/v26/logger"
	"github.com/OrbitOS-org/orbit-os-sdk-go/v26/metadata"
)

const logTag = "hello_web"

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
	// route is the path the Launcher serves the page at: http://<device>/hello_web.
	route = "/hello_web"
	// Web interface ports are reserved to the range portMin–portMax. The app
	// starts at preferredPort and moves to the next one when a port is taken.
	portMin       = 50000
	portMax       = 60000
	preferredPort = 50080
)

//go:embed page.html
var pageHTML []byte

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
	// Read the values shown on the page once, at start-up.
	name, err := client.SystemManager.GetDeviceName()
	if err != nil {
		return fmt.Errorf("GetDeviceName: %w", err)
	}
	model, err := client.SystemManager.GetHardwareModel()
	if err != nil {
		return fmt.Errorf("GetHardwareModel: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(pageHTML)
	})
	mux.HandleFunc("/api/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"deviceName":    name,
			"hardwareModel": model,
			"time":          time.Now().Format(time.RFC3339),
		})
	})

	listener, err := listenAndRegister(client)
	if err != nil {
		return err
	}
	logger.Infof(logTag, "Listening on %s, registered at %s", listener.Addr(), route)

	// Remove the page from the Launcher when the app stops.
	defer func() {
		if err := client.AppHubManager.UnregisterService(); err != nil {
			logger.Errorf(logTag, "UnregisterService: %v", err)
		}
	}()

	server := &http.Server{Handler: mux}
	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(listener) }()

	select {
	case err := <-serveErr:
		return fmt.Errorf("Serve: %w", err)
	case <-ctx.Done():
	}

	// Give the requests in progress a moment to finish.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("Shutdown: %w", err)
	}
	return nil
}

// listenAndRegister finds a free port in the reserved range, starting at
// preferredPort, and registers the page with the Launcher. It returns the
// listener to serve on.
func listenAndRegister(client *orbitos.Client) (net.Listener, error) {
	span := portMax - portMin + 1
	for i := 0; i < span; i++ {
		port := portMin + (preferredPort-portMin+i)%span
		addr := fmt.Sprintf("127.0.0.1:%d", port)

		listener, err := net.Listen("tcp", addr)
		if err != nil {
			continue // port taken: try the next one
		}
		if err := client.AppHubManager.RegisterWebUI(addr, route); err != nil {
			listener.Close() // the Launcher refused this port: try the next one
			logger.Warnf(logTag, "RegisterWebUI %s: %v", addr, err)
			continue
		}
		return listener, nil
	}
	return nil, fmt.Errorf("no free web interface port in %d-%d", portMin, portMax)
}
