package Config



type DriverConfig struct {
	BufferSize int

	RequsetChanLength int
}

func DefaultDriverConfig() DriverConfig {
	return DriverConfig{BufferSize:2048, RequsetChanLength:1024}
}