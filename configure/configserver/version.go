package configserver

import "strings"

type DefaultVersionComparator struct {
}

func (d DefaultVersionComparator) HasUpdate(requestVersion, dataVersion string) bool {
	return strings.Compare(requestVersion, dataVersion) < 0
}
