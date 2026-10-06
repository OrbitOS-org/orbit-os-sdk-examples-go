# Mobile Notification options

Send a push notification with the optional fields: data for the mobile app, a collapse key and a time to live.

An example app for [Orbit OS](https://www.orbit-os.org/?ref=github-sdk-examples-go), built with the [Orbit OS SDK for Go](https://github.com/OrbitOS-org/orbit-os-sdk-go). Part of the [Orbit OS examples for Go](../../README.md).

This example uses a feature added in **API 26.1** and needs a device with API 26.1 or newer. The device must be signed in to an Orbit OS account with at least one phone linked.

## Run

Open **this folder** in VS Code with the [Orbit Studio](https://marketplace.visualstudio.com/items?itemName=orbit-os.orbit-studio) extension. You need **Go 1.25 or newer** and a device with Orbit OS in Developer Mode.

1. In the Orbit sidebar, run **Add / Update SDK** and set your device's IP. Orbit Studio downloads the SDK (`v26.1.0`) into `orbit-os-sdk-go/`.
2. Use **Run** to try it live: the example runs on your computer and talks to the device.
3. Use **Build + Deploy** to install the signed `.orb` on the device.

The project builds against the SDK copy in `orbit-os-sdk-go/`, which is not in the repository: Orbit Studio downloads it. Without Orbit Studio, unpack the [SDK release](https://github.com/OrbitOS-org/orbit-os-sdk-go/releases/tag/v26.1.0) into that folder, then run `go build ./cmd/notification_options`.

## Project layout

| Path | What |
|---|---|
| `cmd/notification_options/main.go` | the app |
| `cmd/notification_options/metadata.json` | manifest: name, version, entry point, permissions |
| `cmd/notification_options/orb/icon.svg` | launcher icon |
| `orbit.project.json` | Orbit Studio project settings |
| `go.mod`, `go.work` | Go module and workspace, using the local SDK copy in `orbit-os-sdk-go/` |

Permissions declared: `MobileNotificationService/Send`.
