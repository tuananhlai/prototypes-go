package main

import (
	"encoding/json"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	googleConfig = &oauth2.Config{
		// Get client iD https://support.google.com/workspacemigrate/answer/9222992?hl=en
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		// In the oauth client screen, click on the `info` icon on the top right of the page
		// to view the secret.
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		// This redirect url must be explicitly allowed in the oauth client screen
		// of google cloud console.
		RedirectURL:  "http://localhost:8080/callback/google",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/login/google", handleLogin(googleConfig))
	mux.HandleFunc("/callback/google", handleCallback(googleConfig, "https://www.googleapis.com/oauth2/v2/userinfo"))

	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}

func handleLogin(config *oauth2.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url := config.AuthCodeURL("random-state-string", oauth2.AccessTypeOffline)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func handleCallback(config *oauth2.Config, userInfoURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("state") != "random-state-string" {
			http.Error(w, "State mismatch", http.StatusBadRequest)
			return
		}

		code := r.FormValue("code")
		token, err := config.Exchange(r.Context(), code)
		if err != nil {
			http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		client := config.Client(r.Context(), token)
		resp, err := client.Get(userInfoURL)
		if err != nil {
			http.Error(w, "failed to get user info: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		w.Header().Set("Content-Type", "application/json")
		var data any
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			http.Error(w, "json decode error", http.StatusInternalServerError)
			return
		}

		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		enc.Encode(data)
	}
}
