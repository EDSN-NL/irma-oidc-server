package authorizationserver

import (
	"crypto/rsa"
	"github.com/ory/fosite"
	"github.com/ory/fosite-example/config"
	"net/http"
	"time"

	"github.com/ory/fosite/compose"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/storage"
	"github.com/ory/fosite/token/jwt"
)

var hasher = fosite.BCrypt{
	WorkFactor: 12,
}

var oauth2 fosite.OAuth2Provider

var issuer string

func RegisterHandlers() {
	// Set up oauth2 endpoints. You could also use gorilla/mux or any other router.
	http.HandleFunc("/oauth2/auth", authEndpoint)
	http.HandleFunc("/oauth2/token", tokenEndpoint)
}

// Because we are using oauth2 and open connect id, we use this little helper to combine the two in one
// variable.
func getStrategy(fositeConfig *compose.Config, hmacSecret []byte, rsaKey *rsa.PrivateKey) compose.CommonStrategy {
	return compose.CommonStrategy{
		// alternatively you could use:
		//  OAuth2Strategy: compose.NewOAuth2JWTStrategy(mustRSAKey())
		CoreStrategy: compose.NewOAuth2HMACStrategy(fositeConfig, hmacSecret),

		// open id connect strategy
		OpenIDConnectTokenStrategy: compose.NewOpenIDConnectStrategy(rsaKey),
	}
}

func getStoreFromConfig(clients []config.OidcClient) (*storage.MemoryStore, error) {

	fositeClients := map[string]*fosite.DefaultClient{}

	for _, v := range clients {
		secret, err := hasher.Hash([]byte(v.Secret))

		if err != nil {
			return nil, err
		}

		fositeClients[v.ID] = &fosite.DefaultClient{
			ID:            v.ID,
			Secret:        secret,
			RedirectURIs:  v.RedirectURIs,
			ResponseTypes: []string{"id_token", "code"},
			GrantTypes:    []string{"authorization_code"},
			Scopes:        v.Scopes,
		}
	}

	return &storage.MemoryStore{
		IDSessions:             make(map[string]fosite.Requester),
		Clients:                fositeClients,
		AuthorizeCodes:         map[string]storage.StoreAuthorizeCode{},
		Implicit:               map[string]fosite.Requester{},
		AccessTokens:           map[string]fosite.Requester{},
		RefreshTokens:          map[string]fosite.Requester{},
		PKCES:                  map[string]fosite.Requester{},
		AccessTokenRequestIDs:  map[string]string{},
		RefreshTokenRequestIDs: map[string]string{},
	}, nil
}

func SetOauth2Provider(config config.IrmaOpenIDServerConfig) error {
	issuer = config.OidcIssuer
	fositeConfig := new(compose.Config)
	strategy := getStrategy(fositeConfig, []byte(config.OauthHmacSecret), config.JwtPrivateKey)

	store, err := getStoreFromConfig(config.OidcClients)
	if err != nil {
		return err
	}

	oauth2 = compose.Compose(
		fositeConfig,
		store,
		strategy,
		&hasher,

		// enabled handlers
		compose.OAuth2AuthorizeExplicitFactory,

		// be aware that open id connect factories need to be added after oauth2 factories to work properly.
		compose.OpenIDConnectExplicitFactory,
	)
	return nil
}

// newSession is a helper function for creating a new session. This may look like a lot of code but since we are
// setting up multiple strategies it is a bit longer.
// Usually, you could do:
//  session = new(fosite.DefaultSession)
func newSession(subject string, disclosed map[string]interface{}) *openid.DefaultSession {
	extra := make(map[string]interface{})
	extra["disclosed"] = disclosed

	return &openid.DefaultSession{
		Claims: &jwt.IDTokenClaims{
			Issuer:      issuer,
			Subject:     subject,
			ExpiresAt:   time.Now().Add(time.Minute * 5),
			IssuedAt:    time.Now(),
			RequestedAt: time.Now(),
			AuthTime:    time.Now(),
			Extra:       extra,
		},
		Headers: &jwt.Headers{},
	}
}
