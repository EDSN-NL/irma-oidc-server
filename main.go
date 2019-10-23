package main

import (
	"fmt"
	"github.com/ory/fosite-example/irma"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/ory/fosite-example/authorizationserver"
	"github.com/ory/fosite-example/oauth2client"
	"github.com/privacybydesign/irmago/server"
	"github.com/privacybydesign/irmago/server/irmaserver"
	goauth "golang.org/x/oauth2"
)

// A valid oauth2 client (check the store) that additionally requests an OpenID Connect id token
var clientConf = goauth.Config{
	ClientID:     "my-client2",
	ClientSecret: "foobar",
	RedirectURL:  "http://localhost:3846/callback",
	Scopes:       []string{"openid"},
	Endpoint: goauth.Endpoint{
		TokenURL: "http://localhost:3846/oauth2/token",
		AuthURL:  "http://localhost:3846/oauth2/auth",
	},
}

func main() {
	// ### IRMA ###
	configuration := &server.Configuration{
		// Replace with address that IRMA apps can reach
		URL: "https://REPLACE_ME
		EnableSSE: true,
	}

	err := irmaserver.Initialize(configuration)
	if err != nil {
		// ...
	}
	http.Handle("/irma/", irmaserver.HandlerFunc())

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/irma-login", irma.CreateSessionRequest)
	http.HandleFunc("/get-irma-session", irma.GetIrmaSessionPtr)

	// ### oauth2 server ###
	authorizationserver.RegisterHandlers() // the authorization server (fosite)

	// ### oauth2 client ###
	// the following handlers are oauth2 consumers
	http.HandleFunc("/callback", oauth2client.CallbackHandler(clientConf)) // the oauth2 callback endpoint

	port := "3846"
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}

	fmt.Println("Please open your webbrowser at http://localhost:" + port)
	_ = exec.Command("open", "http://localhost:"+port).Run()
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
