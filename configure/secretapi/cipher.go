package secretapi

type Level2CipherTool struct {
	storage KeyStorage
}

func NewLevel2CipherTool(storage KeyStorage) *Level2CipherTool {
	return &Level2CipherTool{
		storage: storage,
	}
}
