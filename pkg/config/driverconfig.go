package Config



type DriverConfig struct {
	BufferSize int

	RequsetChanLength int
}

func DefaultDriverConfig() DriverConfig {
	return DriverConfig{BufferSize:1024, RequsetChanLength:1024}
}