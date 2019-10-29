package irma

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-session/session"
	"github.com/gobuffalo/packr"
	"github.com/privacybydesign/irmago"
	"github.com/privacybydesign/irmago/server"
	"github.com/privacybydesign/irmago/server/irmaserver"
	"net/http"
	"net/url"
	"strings"
)

var box = packr.NewBox("./templates")

func GetIrmaSessionPtr(w http.ResponseWriter, r *http.Request) {
	store, err := session.Start(nil, w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	sessionPtr, ok := store.Get("IrmaSession")
	if !ok || sessionPtr == nil {
		http.Error(w, "No IrmaSession Found", http.StatusBadRequest)
		return
	}

	store.Delete("IrmaSession")
	store.Save()
	w.Header().Add("Content-Type", "text/json")
	w.Write(sessionPtr.([]byte))
}

func CreateSessionRequest(w http.ResponseWriter, r *http.Request) {
	// Don't ever cache this page to prevent irma session from being created
	w.Header().Set("Cache-Control", "no-cache")

	store, err := session.Start(nil, w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	oauthSessionData, ok := store.Get("OAuthSessionData")
	if !ok {
		http.Error(w, "no OAuthSessionData present", http.StatusBadRequest)
		return
	}

	var form url.Values
	form = oauthSessionData.(url.Values)
	request := getDisclosureRequestFromScopeParam(form.Get("scope"))

	sessionPointer, _, err := irmaserver.StartSession(request, func(r *server.SessionResult) {
		fmt.Println("IRMA Session done, result: ", server.ToJson(r))
		store.Set("IrmaResult", server.ToJson(r))
		store.Save()
	})

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	jsonSessionPointer, _ := json.Marshal(sessionPointer)
	store.Set("IrmaSession", jsonSessionPointer)
	store.Save()

	outputHTML(w, r, "irma.html")
}

func outputHTML(w http.ResponseWriter, req *http.Request, filename string) {
	file, err := box.Find(filename)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write(file)
}

func SessionResultToMap(sessionResult *server.SessionResult) map[string]interface{} {
	toReturn := make(map[string]interface{})

	for _, v := range sessionResult.Disclosed {
		for _, w := range v {
			toReturn[w.Identifier.String()] = *w.RawValue
		}
	}

	return toReturn
}

func ExtractAttribute(sessionResult *server.SessionResult, id irma.AttributeTypeIdentifier) (string, error) {
	for _, v := range sessionResult.Disclosed {
		for _, w := range v {
			if w.Identifier == id {
				if w.Status != irma.AttributeProofStatusPresent {
					return "", errors.New(fmt.Sprintf("invalid irma proof: %v", w.Identifier))
				}
				return *w.RawValue, nil
			}
		}
	}

	return "", errors.New("not found")
}

func ExtractSubjectIdentifier(scopes []string) (irma.AttributeTypeIdentifier, error) {
	if len(scopes) < 1 {
		return irma.AttributeTypeIdentifier{}, errors.New("invalid scope")
	}

	if scopes[0] != "openid" {
		return irma.NewAttributeTypeIdentifier(scopes[0]), nil
	}

	// Since openid is also included, scopes list should be at least 2
	if len(scopes) < 2 {
		return irma.AttributeTypeIdentifier{}, errors.New("invalid scope")
	}

	// Return the first 'not openid' from scope list
	return irma.NewAttributeTypeIdentifier(scopes[1]), nil
}

func getDisclosureRequestFromScopeParam(scope string) *irma.DisclosureRequest {
	scopes := strings.Split(scope, " ")

	var attributeRequests []irma.AttributeRequest
	for _, v := range scopes {
		if v == "openid" {
			continue
		}
		attributeRequests = append(attributeRequests, irma.NewAttributeRequest(v))
	}

	request := irma.NewDisclosureRequest()
	request.Disclose = irma.AttributeConDisCon{
		irma.AttributeDisCon{
			attributeRequests,
		},
	}
	return request
}
