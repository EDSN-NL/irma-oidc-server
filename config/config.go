package config

type OidcClient struct {
	ID           string
	Secret       string
	RedirectURIs []string
	Scopes       []string
}

type IrmaOpenIDServerConfig struct {
	// TODO: use irma.IrmaConfig here?
	IrmaURL    string
	Port       int
	Clients    []OidcClient
	HmacSecret string
	Issuer     string
}

func GetConfig() IrmaOpenIDServerConfig {
	return IrmaOpenIDServerConfig{
		IrmaURL: "http://TODO",
		Port:    3846,
		Clients: []OidcClient{
			{
				ID:           "my-client2",
				Secret:       "foobar",
				RedirectURIs: []string{"http://localhost:8080"},
				Scopes:       []string{"openid", "pbdf.gemeente.personalData.initials", "pbdf.gemeente.personalData.surname", "pbdf.pbdf.email.email"},
			},
		},
		HmacSecret: "some-super-cool-secret-that-nobody-knows",
		Issuer:     "irma-oidc-server",
	}
}
