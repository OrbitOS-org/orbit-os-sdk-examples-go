# GPIO button and LED

Read a button on an input pin and drive an LED on an output pin with it.

An example app for [Orbit OS](https://www.orbit-os.org/?ref=github-sdk-examples-go), built with the [Orbit OS SDK for Go](https://github.com/OrbitOS-org/orbit-os-sdk-go). Part of the [Orbit OS examples for Go](../../README.md).

The pins used are constants at the top of `main.go`. Change them to match your board and wiring before you run. The button must drive its pin high when pressed (use a pull-down resistor).

## Run

Open **this folder** in VS Code with the [Orbit Studio](https://marketplace.visualstudio.com/items?itemName=orbit-os.orbit-studio) extension. You need **Go 1.25 or newer** and a device with Orbit OS in Developer Mode.

1. In the Orbit sidebar, run **Add / Update SDK** and set your device's IP. Orbit Studio downloads the SDK (`v26.0.3`) into `orbit-os-sdk-go/`.
2. Use **Run** to try it live: the example runs on your computer and talks to the device.
3. Use **Build + Deploy** to install the signed `.orb` on the device.

The project builds against the SDK copy in `orbit-os-sdk-go/`, which is not in the repository: Orbit Studio downloads it. Without Orbit Studio, unpack the [SDK release](https://github.com/OrbitOS-org/orbit-os-sdk-go/releases/tag/v26.0.3) into that folder, then run `go build ./cmd/gpio_button_led`.

## Project layout

| Path | What |
|---|---|
| `cmd/gpio_button_led/main.go` | the app |
| `cmd/gpio_button_led/metadata.json` | manifest: name, version, entry point, permissions |
| `cmd/gpio_button_led/orb/icon.svg` | launcher icon |
| `orbit.project.json` | Orbit Studio project settings |
| `go.mod`, `go.work` | Go module and workspace, using the local SDK copy in `orbit-os-sdk-go/` |

Permissions declared: `SystemService/*`, `GpioService/*`.
