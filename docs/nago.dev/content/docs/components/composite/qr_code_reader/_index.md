---
title: QR Code Reader
---

The QR code reader shows the camera stream and writes every decoded QR code into its state. List the
cameras of the user with `wnd.MediaDevices()` and pass the one to use. Without a camera, the reader shows
the `NoMediaDeviceContent`.

```go
func view(wnd core.Window) core.View {
    cameras := core.AutoState[[]core.MediaDevice](wnd)
    scanned := core.AutoState[[]string](wnd)

    core.OnAppear(wnd, "list-cameras", func(ctx context.Context) {
        wnd.MediaDevices().List(core.MediaDeviceListOptions{WithVideo: true}).Observe(func(devices []core.MediaDevice, err error) {
            if err == nil {
                cameras.Set(devices)
            }
        })
    })

    var camera core.MediaDevice
    if len(cameras.Get()) > 0 {
        camera = cameras.Get()[0]
    }

    return VStack(
        QrCodeReader(camera).
            InputValue(scanned).
            ShowTracker(true).
            NoMediaDeviceContent(Text("No camera available")).
            Frame(Frame{}.Size(L320, L320)),
        Text("Scanned: "+strings.Join(scanned.Get(), ", ")),
    ).Gap(L16)
}
```

## Constructors

```go
func QrCodeReader(mediaDevice core.MediaDevice) TQrCodeReader
```

QrCodeReader creates a new QR code reader using the given media device (camera).

## Methods

| Method | Description |
|--------|-------------|
| `ActivatedTorch(activatedTorch bool) TQrCodeReader` | ActivatedTorch enables or disables the camera torch (flashlight). |
| `Frame(frame Frame) TQrCodeReader` | Frame sets the layout frame for the QR code reader component. |
| `InputValue(inputValue *core.State[[]string]) TQrCodeReader` | InputValue binds the QR code reader to a state, which will be updated with the scanned QR code values. |
| `NoMediaDeviceContent(noMediaDeviceContent core.View) TQrCodeReader` | NoMediaDeviceContent sets the fallback view shown when no media device (camera) is available. |
| `OnCameraReady(onCameraReady func()) TQrCodeReader` | OnCameraReady sets the callback function to be executed when the camera is ready for scanning. |
| `ShowTracker(showTracker bool) TQrCodeReader` | ShowTracker toggles the visibility of the tracker overlay on the camera preview. |
| `TrackerColor(trackerColor Color) TQrCodeReader` | TrackerColor sets the color of the tracker overlay. |
| `TrackerLineWidth(trackerLineWidth int) TQrCodeReader` | TrackerLineWidth sets the thickness of the tracker overlay lines. |

## Related

- [QR Code](../../basic/qr_code/)
- Tutorial [tutorial-64-qrcodereader](/docs/examples/tutorial-64-qrcodereader/)
- Tutorial [tutorial-63-mediadevices](/docs/examples/tutorial-63-mediadevices/)
