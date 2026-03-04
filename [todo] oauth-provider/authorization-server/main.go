package main

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const (
	addr = ":8080"
)

func main() {
	http.HandleFunc("GET /oauth2/authorize", func(w http.ResponseWriter, r *http.Request) {
		authorizeReq, err := parseAuthorizeQuery(r.URL.Query())
		if err != nil {
			http.Error(w, "can not parse queries", http.StatusBadRequest)
			return
		}

		// validate authorization request + generate redirect uri

		http.Redirect(w, r, authorizeReq.RedirectURI.String(), http.StatusFound)
	})

	http.HandleFunc("POST /oauth2/token", func(w http.ResponseWriter, r *http.Request) {

	})

	log.Println("start server on port", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalln("failed to start http server")
	}
}

type getTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type authorizeRequest struct {
	ResponseType        string
	ClientID            string
	RedirectURI         *url.URL
	Scope               []string
	CodeChallenge       string
	CodeChallengeMethod string
}

func parseAuthorizeQuery(queryValues url.Values) (*authorizeRequest, error) {
	redirectURI, err := url.Parse(queryValues.Get("redirect_uri"))
	if err != nil {
		return nil, err
	}

	req := &authorizeRequest{
		ResponseType:        queryValues.Get("response_type"),
		ClientID:            queryValues.Get("client_id"),
		RedirectURI:         redirectURI,
		Scope:               strings.Split(queryValues.Get("scope"), " "),
		CodeChallenge:       queryValues.Get("code_challenge"),
		CodeChallengeMethod: queryValues.Get("code_challenge_method"),
	}

	if err := validateAuthorizeRequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

// validateAuthorizeRequest ...
func validateAuthorizeRequest(a *authorizeRequest) error {
	if a.ResponseType != "code" {
		return errors.New("invalid response type")
	}
	if a.CodeChallengeMethod != "S256" {
		return errors.New("invalid code challenge method")
	}
	if a.ClientID == "" || len(a.Scope) == 0 || a.CodeChallenge == "" {
		return errors.New("required fields missing")
	}

	return nil
}

func createAuthorizeSuccessRedirectURI(url, code, state string) (string, error) {
	return "", nil
}
