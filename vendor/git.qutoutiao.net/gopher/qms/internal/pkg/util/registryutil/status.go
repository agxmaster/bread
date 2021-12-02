package registryutil

type Status string

const (
	Unknown Status = "unknown"
	Online  Status = "online"
	Offline Status = "offline"
)

func (s Status) String() string {
	return string(s)
}

func GetStatus(meta map[string]string) Status {
	switch meta["Status"] {
	case Online.String():
		return Online
	case Offline.String():
		return Offline
	default:
		return Unknown
	}
}
