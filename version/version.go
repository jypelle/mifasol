package version

import "strconv"

var (
	MajorVersionNumber int = 0
	MinorVersionNumber int = 0
	PatchVersionNumber int = 1
)

func Version() string {
	return strconv.Itoa(MajorVersionNumber)+"."+strconv.Itoa(MinorVersionNumber)+"."+strconv.Itoa(PatchVersionNumber)
}
