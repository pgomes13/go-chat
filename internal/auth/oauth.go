package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	cookieStore *sessions.CookieStore
	oauthCfg    *oauth2.Config
)

// User is the authenticated user stored in the session.
type User struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Picture string `json:"picture"`
}

func Init(clientID, clientSecret, redirectURL, sessionSecret string) {
	oauthCfg = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint:     google.Endpoint,
	}
	cookieStore = sessions.NewCookieStore([]byte(sessionSecret))
	cookieStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

// HandleLogin redirects the user to Google's OAuth consent screen.
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	state := randomState()
	sess, _ := cookieStore.Get(r, sessionName)
	sess.Values["oauth_state"] = state
	sess.Save(r, w)
	http.Redirect(w, r, oauthCfg.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

// HandleCallback exchanges the code for a token and stores the user in the session.
func HandleCallback(w http.ResponseWriter, r *http.Request) {
	sess, _ := cookieStore.Get(r, sessionName)

	if r.FormValue("state") != sess.Values["oauth_state"] {
		http.Error(w, "invalid oauth state", http.StatusBadRequest)
		return
	}

	token, err := oauthCfg.Exchange(r.Context(), r.FormValue("code"))
	if err != nil {
		http.Error(w, "token exchange failed", http.StatusInternalServerError)
		return
	}

	gu, err := fetchGoogleUser(r.Context(), token)
	if err != nil {
		http.Error(w, "failed to fetch user info", http.StatusInternalServerError)
		return
	}

	sess.Values["user_id"] = gu.Sub
	sess.Values["user_name"] = gu.Name
	sess.Values["user_email"] = gu.Email
	sess.Values["user_picture"] = gu.Picture
	delete(sess.Values, "oauth_state")
	sess.Save(r, w)

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// HandleLogout clears the session.
func HandleLogout(w http.ResponseWriter, r *http.Request) {
	sess, _ := cookieStore.Get(r, sessionName)
	sess.Options.MaxAge = -1
	sess.Save(r, w)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// HandleMe returns the logged-in user as JSON, or 401 if not authenticated.
func HandleMe(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// GetUser returns the authenticated user from the session, or nil.
func GetUser(r *http.Request) *User {
	sess, err := cookieStore.Get(r, sessionName)
	if err != nil {
		return nil
	}
	id, _ := sess.Values["user_id"].(string)
	if id == "" {
		return nil
	}
	name, _ := sess.Values["user_name"].(string)
	email, _ := sess.Values["user_email"].(string)
	picture, _ := sess.Values["user_picture"].(string)
	return &User{ID: id, Name: name, Email: email, Picture: picture}
}

type googleUserInfo struct {
	Sub     string `json:"sub"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Picture string `json:"picture"`
}

func fetchGoogleUser(ctx context.Context, token *oauth2.Token) (*googleUserInfo, error) {
	client := oauthCfg.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var u googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, err
	}
	return &u, nil
}

func randomState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
