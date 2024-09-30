package configapi

func SelectorsHelperCache(s *Selectors) {
	s.cache()
}

func SelectorsHelperCacheValue(s *Selectors) string {
	s.cache()
	return s.cached
}
