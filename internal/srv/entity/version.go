package entity

import "github.com/jypelle/mifasol/restApiV1"

// Version

type VersionEntity struct {
	MajorNumber int64
	MinorNumber int64
	PatchNumber int64
}

func (e *VersionEntity) Fill(s *restApiV1.Version) {
	s.MajorNumber = e.MajorNumber
	s.MinorNumber = e.MinorNumber
	s.PatchNumber = e.PatchNumber
}

func (e *VersionEntity) LoadMeta(s *restApiV1.Version) {
	if s != nil {
		e.MajorNumber = s.MajorNumber
		e.MinorNumber = s.MinorNumber
		e.PatchNumber = s.PatchNumber
	}
}
