package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type OidcClient struct {
	ID           string   `yaml:"id,omitempty"`
	Secret       string   `yaml:"secret,omitempty"`
	RedirectURIs []string `yaml:"redirectUris,omitempty"`
	Scopes       []string `yaml:"scopes,omitempty"`
}

type IrmaOpenIDServerConfig struct {
	// TODO: use irma.IrmaConfig here?
	IrmaURL             string          `yaml:"irmaUrl,omitempty"`
	Port                int             `yaml:"port,omitempty"`
	OidcClients         []OidcClient    `yaml:"oidcClients,omitempty"`
	OidcIssuer          string          `yaml:"oidcIssuer,omitempty"`
	OauthHmacSecret     string          `yaml:"oauthHmacSecret,omitempty"`
	JwtPrivateKeyString string          `yaml:"jwtPrivateKey,omitempty"`
	JwtPrivateKey       *rsa.PrivateKey `yaml:"jwtPrivateKey2,omitempty"`
}

func (c *OidcClient) String() string {
	return fmt.Sprintf(
		"\n    ClientID: %v\n    ClientSecret: %v\n    RedirectURIs: %v\n    Scopes: %v\n",
		c.ID,
		c.Secret,
		c.RedirectURIs,
		c.Scopes,
	)
}

func (c *IrmaOpenIDServerConfig) String() string {
	clientString := "    ---"
	for _, v := range c.OidcClients {
		clientString = clientString + v.String() + "    ---"
	}

	return fmt.Sprintf(
		"\n  IrmaURL: %v\n  Port: %v\n  OidcIssuer: %v\n  OAuthHmacSecret %v\n  Registered clients: \n%v\n  JWT Public key: \n%v\n",
		c.IrmaURL,
		c.Port,
		c.OidcIssuer,
		c.OauthHmacSecret,
		clientString,
		printPublicKeyAsPem(c.JwtPrivateKey),
	)
}

func getDefaultConfig() *IrmaOpenIDServerConfig {
	return &IrmaOpenIDServerConfig{
		IrmaURL: "http://localhost:3846/irma",
		Port:    3846,
		OidcClients: []OidcClient{
			{
				ID:           "default-client",
				Secret:       "default-secret",
				RedirectURIs: []string{"http://localhost:8080"},
				Scopes:       []string{"openid", "pbdf.pbdf.email.email"},
			},
		},
		OidcIssuer:      "irma-oidc-server",
		OauthHmacSecret: "some-super-cool-secret-that-nobody-knows",
		JwtPrivateKey:   genRSAKey(),
	}
}

func readConfig() (*IrmaOpenIDServerConfig, error) {
	configFile, err := ioutil.ReadFile("config.yaml")
	parsedConfig := &IrmaOpenIDServerConfig{}
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(configFile, parsedConfig)

	if err != nil {
		return nil, err
	}

	if parsedConfig.JwtPrivateKeyString == "" {
		fmt.Println("Warning: No private key set for JWT signing, generating a new one")
		parsedConfig.JwtPrivateKey = genRSAKey()
		return parsedConfig, nil
	}

	block, _ := pem.Decode([]byte(parsedConfig.JwtPrivateKeyString))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	parsedConfig.JwtPrivateKey = key

	return parsedConfig, nil
}

func GetConfig() *IrmaOpenIDServerConfig {
	parsedConfig, err := readConfig()
	defaultConfig := getDefaultConfig()

	if err != nil {
		fmt.Printf("Warning, cannot parse config.yaml: %v\n", err.Error())
		fmt.Println("Falling back to default config")
		return defaultConfig
	}

	// Fill empty fields with defaults
	if parsedConfig.Port == 0 {
		parsedConfig.Port = defaultConfig.Port
	}
	if parsedConfig.IrmaURL == "" {
		parsedConfig.IrmaURL = defaultConfig.IrmaURL
	}
	if parsedConfig.OidcIssuer == "" {
		parsedConfig.OidcIssuer = defaultConfig.OidcIssuer
	}
	if parsedConfig.OauthHmacSecret == "" {
		parsedConfig.OauthHmacSecret = defaultConfig.OauthHmacSecret
	}
	if len(parsedConfig.OidcClients) == 0 {
		parsedConfig.OidcClients = defaultConfig.OidcClients
	}

	return parsedConfig
}

func genRSAKey() *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	return key
}

func printPublicKeyAsPem(privkey *rsa.PrivateKey) string {
	pubkey := privkey.Public().(*rsa.PublicKey)

	pubkey_bytes, err := x509.MarshalPKIXPublicKey(pubkey)
	if err != nil {
		return ""
	}
	pubkey_pem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: pubkey_bytes,
		},
	)

	return string(pubkey_pem)
}
