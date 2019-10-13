package version

import "strconv"

var (
	MajorVersionNumber int = 0
	MinorVersionNumber int = 1
	PatchVersionNumber int = 0
)

func Version() string {
	return strconv.Itoa(MajorVersionNumber) + "." + strconv.Itoa(MinorVersionNumber) + "." + strconv.Itoa(PatchVersionNumber)
}
