package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	url2 "net/url"
)

type SecretConfig struct {
	Web SecretConfigWeb `json:"web"`
}

type SecretConfigWeb struct {
	ClientId        string `json:"client_id"`
	ProjectId       string `json:"project_id"`
	AuthUri         string `json:"auth_uri"`
	TokenUri        string `json:"token_uri"`
	AuthProviderUrl string `json:"auth_provider_x509_cert_url"`
	ClientSecret    string `json:"client_secret"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

func (tr TokenResponse) String() string {
	return fmt.Sprintf(`==TOKEN RESPONSE==
access_token: %s
token_type: %s
expired_in: %d
refresh_token: %s
`, tr.AccessToken, tr.TokenType, tr.ExpiresIn, tr.RefreshToken)
}

func (secretConfig SecretConfig) createAuthUrl(scopeUrl string) *url2.URL {
	url := &url2.URL{}
	url.Scheme = "https"
	url.Host = "accounts.google.com"
	url.Path = "o/oauth2/auth"

	query := url.Query()
	query.Set("client_id", secretConfig.Web.ClientId)
	query.Set("redirect_uri", "http://localhost:8080")
	query.Set("scope", scopeUrl)
	query.Set("response_type", "code")
	query.Set("approval_prompt", "force")
	query.Set("access_type", "offline")

	url.RawQuery = query.Encode()

	return url
}

func (secretConfig SecretConfig) getToken(code string) *TokenResponse {
	url := &url2.URL{}
	url.Scheme = "https"
	url.Host = "accounts.google.com"
	url.Path = "o/oauth2/token"

	values := url2.Values{}
	values.Add("client_id", secretConfig.Web.ClientId)
	values.Add("client_secret", secretConfig.Web.ClientSecret)
	values.Add("redirect_uri", "http://localhost:8080")
	values.Add("grant_type", "authorization_code")
	values.Add("code", code)

	res, err := http.PostForm(url.String(), values)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	bodyByte, _ := ioutil.ReadAll(res.Body)
	tokenResponse := new(TokenResponse)
	err = json.Unmarshal(bodyByte, tokenResponse)
	if err != nil {
		log.Fatal(err)
	}
	return tokenResponse
}

func NewSecretConfig(jsonPath string) *SecretConfig {
	jsonFile, err := ioutil.ReadFile(jsonPath)
	if err != nil {
		log.Fatal(err)
	}
	secretConfig := new(SecretConfig)
	json.Unmarshal(jsonFile, &secretConfig)
	return secretConfig
}

func httpServer(authCodeCh chan<- string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		queries := r.URL.Query()
		if queries == nil {
			return
		}
		for key, value := range queries {
			fmt.Fprintf(w, "%s: %s\n", key, value)
			if key == "code" {
				fmt.Println("get code.")
				authCodeCh <- value[0]
			}
		}
	})
	http.ListenAndServe(":8080", nil)
}

func main() {
	var (
		jsonPath string
		scopeUrl string
	)
	flag.StringVar(&jsonPath, "json", "", "path of json, which has client_id, client_secret, etc.")
	flag.StringVar(&scopeUrl, "scope", "", "scope url of google services.")
	flag.Parse()

	if jsonPath == "" {
		log.Fatal("You have to set -json option.")
	}
	if scopeUrl == "" {
		log.Fatal("You have to set -scope option.")
	}

	secretConfig := NewSecretConfig(jsonPath)

	authCodeCh := make(chan string)
	go httpServer(authCodeCh)

	url := secretConfig.createAuthUrl(scopeUrl)
	fmt.Println("Click here and authorize your account.")
	fmt.Println(url)

	var authCode string
	for {
		authCode = <-authCodeCh
		break
	}
	tokenResponse := secretConfig.getToken(authCode)
	fmt.Println(tokenResponse)
}
