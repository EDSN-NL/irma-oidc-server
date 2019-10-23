package authorizationserver

import (
	"crypto/rand"
	"crypto/rsa"
	"github.com/ory/fosite"
	"net/http"
	"time"

	"github.com/ory/fosite/compose"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/storage"
	"github.com/ory/fosite/token/jwt"
)

func RegisterHandlers() {
	// Set up oauth2 endpoints. You could also use gorilla/mux or any other router.
	http.HandleFunc("/oauth2/auth", authEndpoint)
	http.HandleFunc("/oauth2/token", tokenEndpoint)
}

// This is an exemplary storage instance. We will add a client and a user to it so we can use these later on.
var store = &storage.MemoryStore{
	IDSessions: make(map[string]fosite.Requester),
	Clients: map[string]*fosite.DefaultClient{
		"my-client2": {
			ID:            "my-client2",
			Secret:        []byte(`$2a$10$IxMdI6d.LIRZPpSfEwNoeu4rY3FhDREsxFJXikcgdRRAStxUlsuEO`), // = "foobar"
			RedirectURIs:  []string{"http://localhost:3847/callback"},
			ResponseTypes: []string{"id_token", "code"}, // token not needed?
			GrantTypes:    []string{"authorization_code"},
			Scopes:        []string{"openid", "pbdf.gemeente.personalData.initials", "pbdf.gemeente.personalData.surname", "pbdf.pbdf.email.email"},
		},
	},
	AuthorizeCodes:         map[string]storage.StoreAuthorizeCode{},
	Implicit:               map[string]fosite.Requester{},
	AccessTokens:           map[string]fosite.Requester{},
	RefreshTokens:          map[string]fosite.Requester{},
	PKCES:                  map[string]fosite.Requester{},
	AccessTokenRequestIDs:  map[string]string{},
	RefreshTokenRequestIDs: map[string]string{},
}

var config = new(compose.Config)

// Because we are using oauth2 and open connect id, we use this little helper to combine the two in one
// variable.
var strat = compose.CommonStrategy{
	// alternatively you could use:
	//  OAuth2Strategy: compose.NewOAuth2JWTStrategy(mustRSAKey())
	CoreStrategy: compose.NewOAuth2HMACStrategy(config, []byte("some-super-cool-secret-that-nobody-knows")),

	// open id connect strategy
	OpenIDConnectTokenStrategy: compose.NewOpenIDConnectStrategy(mustRSAKey()),
}

//var oauth2 = compose.ComposeAllEnabled(config, store, )
var oauth2 = compose.Compose(
	config,
	store,
	strat,
	nil,

	// enabled handlers
	compose.OAuth2AuthorizeExplicitFactory,

	// be aware that open id connect factories need to be added after oauth2 factories to work properly.
	compose.OpenIDConnectExplicitFactory,
)

// newSession is a helper function for creating a new session. This may look like a lot of code but since we are
// setting up multiple strategies it is a bit longer.
// Usually, you could do:
//  session = new(fosite.DefaultSession)
func newSession(subject string, disclosed map[string]interface{}) *openid.DefaultSession {
	extra := make(map[string]interface{})
	extra["disclosed"] = disclosed

	return &openid.DefaultSession{
		Claims: &jwt.IDTokenClaims{
			Issuer:  "https://fosite.my-application.com",
			Subject: subject,
			//Audience:    []string{"https://my-client.my-application.com"},
			ExpiresAt:   time.Now().Add(time.Minute * 5),
			IssuedAt:    time.Now(),
			RequestedAt: time.Now(),
			AuthTime:    time.Now(),
			Extra:       extra,
		},
		Headers: &jwt.Headers{},
	}
}

func mustRSAKey() *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	return key
}
