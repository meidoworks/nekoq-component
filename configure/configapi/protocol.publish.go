package configapi

type PublishReq struct {
	Configuration     RawConfiguration `cbor:"cfg,"`
	Selectors         Selectors        `cbor:"selectors,"`
	OptionalSelectors Selectors        `cbor:"opt_selectors,"`
	Operator          string           `cbor:"operator,"`
	Description       string           `cbor:"description"`
}

type PublishRes struct {
	Success bool   `cbor:"success,"`
	Code    string `cbor:"code,"`
	Message string `cbor:"message,"`
}
