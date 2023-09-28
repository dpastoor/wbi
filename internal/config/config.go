package config

type OperatingSystem int

const (
	Unknown OperatingSystem = iota
	Ubuntu20
	Ubuntu22
	Redhat7
	Redhat8
	Redhat9
)

func (os OperatingSystem) ToString() string {
	switch os {
	case Ubuntu20:
		return "Ubuntu 20"
	case Ubuntu22:
		return "Ubuntu 22"
	case Redhat7:
		return "RHEL 7"
	case Redhat8:
		return "RHEL 8"
	case Redhat9:
		return "RHEL 9"
	default:
		return "Unknown"
	}
}
