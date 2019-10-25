package config

type IrmaOpenIDServerConfig struct {
	// TODO: use irma.IrmaConfig here
	IrmaURL string

	Port int
}

func GetConfig() IrmaOpenIDServerConfig {
	return IrmaOpenIDServerConfig{
		IrmaURL: "http://TODO",
		Port:    3846,
	}
}
