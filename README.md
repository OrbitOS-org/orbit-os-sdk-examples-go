<p align="center">
  <img src="https://www.orbit-os.org/images/vscode/orbit-os-logo.png" width="300" alt="Orbit OS">
</p>

<h1 align="center">Orbit OS examples for Go</h1>

<p align="center"><b>Small, single-file apps that show how to use the Orbit OS API from Go — one feature per example.</b></p>

This repository is a collection of example apps for [Orbit OS](https://www.orbit-os.org/?ref=github-sdk-examples-go), built with the [Orbit OS SDK for Go](https://github.com/OrbitOS-org/orbit-os-sdk-go). **Each example folder is a separate [Orbit Studio](https://marketplace.visualstudio.com/items?itemName=orbit-os.orbit-studio) project**, with the same layout as a project created in Orbit Studio: open one folder, run it, deploy it.

Each example is a complete app with its own `main.go`, `metadata.json` manifest and project settings. Read one, run it, then copy the folder as the start of your own app.

Runs on Raspberry Pi, Arduino UNO Q and other ARM64 devices with Orbit OS (free Community Edition).

## Examples

The examples are grouped in folders by topic.

### Getting started — [`getting-started/`](getting-started)

Start here.

| Example | What it shows | Services used |
|---|---|---|
| [`template`](getting-started/template) | The smallest possible app. Copy it to start a new one | System |
| [`basic`](getting-started/basic) | Connect to the device and read system information: API version, device identity, hardware, OS | System |

### System — [`system/`](system)

Information about the device and its state.

| Example | What it shows | Services used |
|---|---|---|
| [`metrics_info`](system/metrics_info) | Read one snapshot of CPU, memory, temperature, disk and network metrics | System |
| [`device_certification`](system/device_certification) | Check whether the app runs on certified hardware and read the certification serial | System |

### Peripherals — [`peripherals/`](peripherals)

Pins, buses and serial ports.

| Example | What it shows | Services used |
|---|---|---|
| [`gpio_output`](peripherals/gpio_output) | List the GPIO lines and blink the first one | GPIO |
| [`gpio_input`](peripherals/gpio_input) | Wait for an input pin (`GPIO24`) to go high | GPIO |
| [`pwm`](peripherals/pwm) | List the PWM channels and set frequency and duty cycle on the first one | PWM |
| [`i2c`](peripherals/i2c) | List the I2C buses, read the configuration and scan for devices | I2C |
| [`spi`](peripherals/spi) | List the SPI devices, read the configuration and run a full-duplex transfer | SPI |
| [`uart`](peripherals/uart) | Open a serial port and print what it receives | UART |
| [`gpio_button_led`](peripherals/gpio_button_led) | Read a button on an input pin and drive an LED on an output pin with it | GPIO |
| [`pwm_fade`](peripherals/pwm_fade) | Raise and lower the duty cycle step by step, like an LED fading in and out, then stop the output | PWM |

### Network — [`network/`](network)

Network interfaces of the device.

| Example | What it shows | Services used |
|---|---|---|
| [`wifi_scan`](network/wifi_scan) | List the WiFi interfaces and scan for networks | WiFi |
| [`eth_status`](network/eth_status) | List the Ethernet interfaces with link state and IP configuration | Ethernet |
| [`wifi_status`](network/wifi_status) | Show, for each WiFi interface, whether it is connected, its mode and the details of the connection or access point | WiFi |

### Notifications — [`notifications/`](notifications)

Push notifications to the owner's phone. Needs internet access and an Orbit OS account on the device.

| Example | What it shows | Services used |
|---|---|---|
| [`mobile_notification`](notifications/mobile_notification) | Send a push notification to the phones linked to the device owner's account | Mobile notifications |
| [`notification_options`](notifications/notification_options) | Send a push notification with the optional fields: data for the mobile app, a collapse key and a time to live | Mobile notifications |

### Media and AI — [`media-ai/`](media-ai)

Cameras. Examples of on-device inference will be added here.

| Example | What it shows | Services used |
|---|---|---|
| [`camera_snapshot`](media-ai/camera_snapshot) | List the cameras, take one picture with the first one and save it to a file | Camera |

### Web interface — [`web-ui/`](web-ui)

Apps with a page in the Orbit OS Launcher.

| Example | What it shows | Services used |
|---|---|---|
| [`hello_web`](web-ui/hello_web) | Serve a web page through the Orbit OS Launcher, behind the device login | System, AppHub |

`mobile_notification`, `notification_options` and `device_certification` use features added in **API 26.1**. They need a device with API 26.1 or newer; the other examples also run on API 26.0.

The hardware examples (`gpio_*`, `pwm*`, `i2c`, `spi`, `uart`, `camera_snapshot`) use the pins, buses and ports set as constants at the top of each `main.go`. Change them to match your board and wiring before you run.

## Run an example

**Recommended: [Orbit Studio](https://marketplace.visualstudio.com/items?itemName=orbit-os.orbit-studio) (VS Code).**

You need [VS Code](https://code.visualstudio.com/) with the Orbit Studio extension, **[Go](https://go.dev/dl/) 1.25 or newer** installed (`go` on your PATH), and a device with Orbit OS in Developer Mode.

1. Clone the repository and open **the folder of one example** in VS Code (not the repository root):
   ```bash
   git clone https://github.com/OrbitOS-org/orbit-os-sdk-examples-go
   code orbit-os-sdk-examples-go/peripherals/gpio_output
   ```
2. In the Orbit sidebar, run **Add / Update SDK** and set your device's IP. Orbit Studio downloads the SDK into `orbit-os-sdk-go/`.
3. Use **Run** to try it live: the example runs on your computer and talks to the device.
4. Use **Build + Deploy** to install the signed `.orb` on the device.

**Without Orbit Studio:** each example builds against the SDK copy in its `orbit-os-sdk-go/` folder, which is not in the repository (Orbit Studio downloads it). Unpack the [SDK release](https://github.com/OrbitOS-org/orbit-os-sdk-go/releases/tag/v26.1.0) into that folder, then:

```bash
cd orbit-os-sdk-examples-go/getting-started/basic
go build ./cmd/basic
```

Use Orbit Studio to package and sign the `.orb`.

## How an example is built

```
<topic>/<example>/cmd/<example>/main.go         the app
<topic>/<example>/cmd/<example>/metadata.json   manifest: name, version, entry point, permissions
<topic>/<example>/cmd/<example>/orb/icon.svg    launcher icon
<topic>/<example>/cmd/certs/grpc/               development certificates (added by Orbit Studio, not committed)
<topic>/<example>/orbit.project.json            Orbit Studio project settings
<topic>/<example>/go.mod, go.work               Go module and workspace, using the local SDK copy
<topic>/<example>/orbit-os-sdk-go/              the SDK (downloaded by Orbit Studio, not committed)
```

- **One client.** Every example creates a client with `orbitos.NewClientAuto(host)` and closes it at the end. On the device it connects locally. The `-host` flag is only used in development, when the example runs on your computer.
- **Permissions.** `metadata.json` lists the services the app may call, such as `GpioService/*`. Orbit OS refuses calls to any service that is not listed, so each example declares only what it uses.
- **SDK version.** `sdkVersion` in each example's `orbit.project.json` is `v26.1.0`: that is the SDK Orbit Studio downloads into `orbit-os-sdk-go/`, and `go.mod` points to that local copy.

## Write your own app

Copy the [`getting-started/template`](getting-started/template) folder, rename it and its `cmd/template` folder, and update `go.mod`, `orbit.project.json` and `metadata.json` (`name`, `slug`, `entry_point`, `package_id`, `permissions`) to the new name. Then add the calls you need from the example closest to what you want to build.

For complete apps with a web interface, see the app repositories of the [Orbit OS organization](https://github.com/OrbitOS-org).

## Links

[Orbit OS](https://www.orbit-os.org/?ref=github-sdk-examples-go) · [Getting started](https://www.orbit-os.org/getting_started.html?ref=github-sdk-examples-go) · [SDK reference](https://www.orbit-os.org/api-reference.html?ref=github-sdk-examples-go) · [Orbit OS SDK for Go](https://github.com/OrbitOS-org/orbit-os-sdk-go) · info@orbit-os.org

## License

Apache-2.0 — see [LICENSE](LICENSE) and [NOTICE](NOTICE).
