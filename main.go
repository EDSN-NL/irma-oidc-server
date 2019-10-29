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
	config := config.GetConfig()
	authorizationserver.SetOauth2Provider(config)

	// ### IRMA ###
	err := irmaserver.Initialize(&server.Configuration{
		// Replace with address that IRMA apps can reach
		URL:       config.IrmaURL,
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

	fmt.Printf("Please open your webbrowser at http://localhost%v\n", config.Port)

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.Port), nil))
}
