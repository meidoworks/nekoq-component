package configapi

type AcquireConfigurationReq struct {
	Requested         []RequestedConfigurationKey `cbor:"req,"`
	Selectors         Selectors                   `cbor:"selectors,"`
	OptionalSelectors Selectors                   `cbor:"opt_selectors,"`
}

type AcquireConfigurationRes struct {
	Requested []Configuration `cbor:"req,"`
}

type AcquireConfigurationFailRes struct {
	Code     string   `cbor:"code,"`
	Message  string   `cbor:"msg,"`
	InfoList []string `cbor:"info_list,"`
}

type GetConfigurationRes struct {
	Code          string        `cbor:"code,"`
	Message       string        `cbor:"msg,"`
	Configuration Configuration `cbor:"cfg"`
}
