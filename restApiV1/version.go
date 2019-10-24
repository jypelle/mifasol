package restApiV1

import "strconv"

type Version struct {
	MajorNumber int64
	MinorNumber int64
	PatchNumber int64
}

func (m *Version) String() string {
	return strconv.FormatInt(m.MajorNumber, 10) + "." + strconv.FormatInt(m.MinorNumber, 10) + "." + strconv.FormatInt(m.PatchNumber, 10)
}

func (m *Version) LowerThan(s Version) bool {
	if m.MajorNumber < s.MajorNumber {
		return true
	}
	if m.MajorNumber > s.MajorNumber {
		return false
	}
	if m.MinorNumber < s.MinorNumber {
		return true
	}
	if m.MinorNumber > s.MinorNumber {
		return false
	}
	if m.PatchNumber < s.PatchNumber {
		return true
	}
	return false
}
