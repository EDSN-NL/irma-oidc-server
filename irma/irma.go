package irma

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/privacybydesign/irmago"
	"net/http"
	"github.com/go-session/session"
	"github.com/privacybydesign/irmago/server"
	"github.com/privacybydesign/irmago/server/irmaserver"
	"os"
)

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


func CreateFullnameRequest(w http.ResponseWriter, r *http.Request) {
	store, err := session.Start(nil, w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	request := `{
      "type": "disclosing",
      "content": [{ "label": "Email", "attributes": [ "pbdf.pbdf.email.email" ]}]
  }`

	sessionPointer, _, err := irmaserver.StartSession(request, func (r *server.SessionResult) {
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
	file, err := os.Open(filename)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer file.Close()
	fi, _ := file.Stat()
	http.ServeContent(w, req, file.Name(), fi.ModTime(), file)
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

