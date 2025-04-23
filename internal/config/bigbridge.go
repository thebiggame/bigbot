package config

// Type BridgeConfig defines the configuration available to BIGbridge.
type BridgeConfig struct {
	WsAddress string       `long:"addr" description:"BIGbot address and port" default:"ws://localhost:8080/ws" env:"ADDR"`
	Key       SecretString `long:"key" description:"BIGbot authentication key" required:"" env:"KEY"`
	AV        struct {
		OBS struct {
			Hostname string       `long:"host" help:"OBS Host" default:"" env:"HOST"`
			Password SecretString `long:"password" help:"OBS password" default:"" env:"PASSWORD"`
		} `prefix:"obs." embed:"" envprefix:"OBS_"`
		NodeCG struct {
			Hostname          string       `long:"host" help:"NodeCG Host" default:"" env:"HOST"`
			BundleName        string       `long:"bundle" help:"NodeCG bundle name" default:"thebiggame" env:"BUNDLE"`
			AuthenticationKey SecretString `long:"key" help:"Authentication key" default:"" env:"AUTHKEY"`
		} `prefix:"nodecg." embed:"" envprefix:"NODECG_"`
	} `prefix:"av." embed:"" envprefix:"AV_"`
}
