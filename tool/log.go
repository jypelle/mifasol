package tool

import (
	"github.com/sirupsen/logrus"
	"time"
)

func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	logrus.Debugf("%s took %s", name, elapsed)
}
