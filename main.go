package main

import (
	"fmt"
	"github.com/ory/fosite-example/authorizationserver"
	"github.com/ory/fosite-example/config"
	"github.com/ory/fosite-example/irma"
	"github.com/privacybydesign/irmago/server"
	"github.com/privacybydesign/irmago/server/irmaserver"
	"log"
	"net/http"
	"strconv"
)

func main() {
	c := config.GetConfig()

	fmt.Printf("Running with config: %v\n", c)

	authorizationserver.SetOauth2Provider(*c)

	// ### IRMA ###
	err := irmaserver.Initialize(&server.Configuration{
		// Replace with address that IRMA apps can reach
		URL:       c.IrmaURL,
		EnableSSE: true,
	})
	if err != nil {
		panic(fmt.Sprintf("Error starting Irma server: %v", err))
	}
	http.Handle("/irma/", irmaserver.HandlerFunc())

	// ### register all OAuth handlers ###
	authorizationserver.RegisterHandlers()

	// ### Other Handlers ###
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/irma-login", irma.CreateSessionRequest)
	http.HandleFunc("/get-irma-session", irma.GetIrmaSessionPtr)

	fmt.Printf("Listening at http://localhost:%v\n", c.Port)

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(c.Port), nil))
}
