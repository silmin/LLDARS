package lldars

const (
	NormalMode LLDARSServeMode = iota
	RevivalMode
)

func (m LLDARSServeMode) String() string {
	return toServeModeString[m]
}

var toServeModeString = map[LLDARSServeMode]string{
	NormalMode:  "NormalMode",
	RevivalMode: "RevivalMode",
}
