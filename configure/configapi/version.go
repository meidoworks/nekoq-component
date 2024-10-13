package configapi

type VersionComparator interface {
	HasUpdate(requestVersion, dataVersion string) bool
}
