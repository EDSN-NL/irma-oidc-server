package authorizationserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-session/session"
	"github.com/ory/fosite"
	"github.com/ory/fosite-example/irma"
	"github.com/privacybydesign/irmago/server"
	"log"
	"net/http"
	"net/url"
)

func authEndpoint(rw http.ResponseWriter, req *http.Request) {
	// This context will be passed to all methods.
	ctx := fosite.NewContext()

	// retrieve irma session info
	store, err := session.Start(nil, rw, req)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if v, ok := store.Get("OAuthSessionData"); ok {
		var form url.Values
		form = v.(url.Values)
		req.Form = form
	}

	irmaResult, ok := store.Get("IrmaResult")
	if !ok {
		// First validate the request
		ar, err := oauth2.NewAuthorizeRequest(ctx, req)
		if err != nil {
			log.Printf("Error Invalid authorize request: %+v", err)
			oauth2.WriteAuthorizeError(rw, ar, err)
			return
		}

		// Save OAuth params in session
		if req.Form == nil {
			req.ParseForm()
		}
		store.Set("OAuthSessionData", req.Form)
		store.Save()

		rw.Header().Set("Location", "/irma-login")
		rw.WriteHeader(http.StatusFound)
		return
	}

	store.Flush()

	irmaParsedResult := server.SessionResult{}
	err = json.Unmarshal([]byte(fmt.Sprintf("%v", irmaResult)), &irmaParsedResult)
	if err != nil {
		fmt.Printf("Could not parse irma session result: %v\n", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	// Let's create an AuthorizeRequest object!
	ar, err := oauth2.NewAuthorizeRequest(ctx, req)
	if err != nil {
		log.Printf("Error occurred in NewAuthorizeRequest: %+v", err)
		oauth2.WriteAuthorizeError(rw, ar, err)
		return
	}

	// You have now access to authorizeRequest, Code ResponseTypes, Scopes ...
	scopes := ar.GetRequestedScopes()
	if len(scopes) < 1 {
		log.Println("Error: No scopes supplied")
		oauth2.WriteAuthorizeError(rw, ar, errors.New("No scopes supplied"))
		return
	}

	subjectId, err := irma.ExtractSubjectIdentifier(scopes)
	if err != nil {
		log.Printf("Invalid scope")
		oauth2.WriteAuthorizeError(rw, ar, err)
		return
	}

	subject, err := irma.ExtractAttribute(&irmaParsedResult, subjectId)
	if err != nil {
		log.Printf("Cannot find IRMA attribute in result %v\n", err)
		oauth2.WriteAuthorizeError(rw, ar, err)
		return
	}

	// Normally, this would be the place where you would check if the user is logged in and gives his consent.
	// We're simplifying things and just checking if the request includes a valid username and password
	// Now that the user is authorized, we set up a session:
	mySessionData := newSession(subject, irma.SessionResultToMap(&irmaParsedResult))
	ar.GrantScope("openid")
	response, err := oauth2.NewAuthorizeResponse(ctx, ar, mySessionData)

	// Catch any errors, e.g.:
	// * unknown client
	// * invalid redirect
	// * ...
	if err != nil {
		log.Printf("Error occurred in NewAuthorizeResponse: %+v", err)
		oauth2.WriteAuthorizeError(rw, ar, err)
		return
	}

	// Last but not least, send the response!
	oauth2.WriteAuthorizeResponse(rw, ar, response)
}
