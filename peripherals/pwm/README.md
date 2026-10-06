# PWM example

List the PWM channels and set frequency and duty cycle on the first one.

An example app for [Orbit OS](https://www.orbit-os.org/?ref=github-sdk-examples-go), built with the [Orbit OS SDK for Go](https://github.com/OrbitOS-org/orbit-os-sdk-go). Part of the [Orbit OS examples for Go](../../README.md).

The pins, buses or ports used are constants at the top of `main.go`. Change them to match your board and wiring before you run.

## Run

Open **this folder** in VS Code with the [Orbit Studio](https://marketplace.visualstudio.com/items?itemName=orbit-os.orbit-studio) extension. You need **Go 1.25 or newer** and a device with Orbit OS in Developer Mode.

1. In the Orbit sidebar, run **Add / Update SDK** and set your device's IP. Orbit Studio downloads the SDK (`v26.0.3`) into `orbit-os-sdk-go/`.
2. Use **Run** to try it live: the example runs on your computer and talks to the device.
3. Use **Build + Deploy** to install the signed `.orb` on the device.

The project builds against the SDK copy in `orbit-os-sdk-go/`, which is not in the repository: Orbit Studio downloads it. Without Orbit Studio, unpack the [SDK release](https://github.com/OrbitOS-org/orbit-os-sdk-go/releases/tag/v26.0.3) into that folder, then run `go build ./cmd/pwm`.

## Project layout

| Path | What |
|---|---|
| `cmd/pwm/main.go` | the app |
| `cmd/pwm/metadata.json` | manifest: name, version, entry point, permissions |
| `cmd/pwm/orb/icon.svg` | launcher icon |
| `orbit.project.json` | Orbit Studio project settings |
| `go.mod`, `go.work` | Go module and workspace, using the local SDK copy in `orbit-os-sdk-go/` |

Permissions declared: `SystemService/*`, `PwmService/*`.
