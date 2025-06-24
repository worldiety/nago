package xmediadevice

type MediaDeviceKind uint64

const (
	AudioInput  MediaDeviceKind = 0
	AudioOutput MediaDeviceKind = 1
	VideoInput  MediaDeviceKind = 2
)

func (k MediaDeviceKind) String() string {
	switch k {
	case AudioInput:
		return "Audio-Input"
	case AudioOutput:
		return "Audio-Output"
	case VideoInput:
		return "Video-Input"
	}
	return "Unbekannt"
}

type MediaDevice struct {
	DeviceID string
	GroupID  string
	Label    string
	Kind     MediaDeviceKind
}
