---
# Content is auto generated
# Manual changes will be overwritten!
title: QR Code Reader
---
It uses a media device (e. g. , camera) to scan QR codes in real-time. The component supports visual trackers, torch (flashlight) activation,
custom UI when no media device is available, and a callback when the camera is ready.

## Constructors
### QrCodeReader
QrCodeReader creates a new QR code reader using the given media device (camera).
By default, it enables the tracker, sets the tracker color to M0, line width to 2,
torch off, and initializes an empty onCameraReady callback.

---
## Methods
| Method | Description |
|--------| ------------|
| `ActivatedTorch(activatedTorch bool)` | ActivatedTorch enables or disables the camera torch (flashlight). |
| `Frame(frame Frame)` | Frame sets the layout frame for the QR code reader component. |
| `InputValue(inputValue *core.State[[]string])` | InputValue binds the QR code reader to a state, which will be updated with the scanned QR code values. |
| `NoMediaDeviceContent(noMediaDeviceContent core.View)` | NoMediaDeviceContent sets the fallback view shown when no media device (camera) is available. |
| `OnCameraReady(onCameraReady func())` | OnCameraReady sets the callback function to be executed when the camera is ready for scanning. |
| `ShowTracker(showTracker bool)` | ShowTracker toggles the visibility of the tracker overlay on the camera preview. |
| `TrackerColor(trackerColor Color)` | TrackerColor sets the color of the tracker overlay. |
| `TrackerLineWidth(trackerLineWidth int)` | TrackerLineWidth sets the thickness of the tracker overlay lines. |
---

## Related
- [Frame](../../layout/frame/)

