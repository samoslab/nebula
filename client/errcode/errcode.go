package errcode

type Status int

const (
	RetSuccess Status = iota
	RetNotFoundConfig
	RetVerifyConfigFailed
	RetSignFailed
	RetTrackerFailed
	RetUnknown
)

var statusString = []string{
	RetSuccess:            "got_reward",
	RetNotFoundConfig:     "not found config file",
	RetVerifyConfigFailed: "verify config file failed",
	RetSignFailed:         "signature failed",
	RetTrackerFailed:      "tracker failed",
	RetUnknown:            "unknown",
}

func (s Status) String() string {
	return statusString[s]
}
