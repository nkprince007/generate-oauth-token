package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/gitlab"
)

func doGitlabOAuthDance(done chan<- int, code <-chan string) {
	var clientID string
	var clientSecret string

	scopes := []string{"api"}
	ctx := context.Background()

	fmt.Println("Create your OAuth application here " +
		"(https://gitlab.com/profile/applications)")
	fmt.Print("Enter your client ID please: ")
	if _, err := fmt.Scan(&clientID); err != nil {
		log.Fatal(err)
	}

	fmt.Print("Enter your client secret please: ")
	if _, err := fmt.Scan(&clientSecret); err != nil {
		log.Fatal(err)
	}

	fmt.Print("Enter your OAuth scopes please (seperated by space): ")
	for {
		var scope string
		if _, err := fmt.Scanln(&scope); err != nil {
			break
		}
		if scope != "" {
			scopes = append(scopes, scope)
		}
	}

	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       scopes,
		RedirectURL:  "http://localhost:8000",
		Endpoint:     gitlab.Endpoint,
	}

	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	fmt.Println("\nVisit the following URL for the auth dialog:")
	fmt.Printf("%s\n", url)
	fmt.Println("Awaiting authentication...")
	openBrowser(url)

	// Use the authorization code that is pushed to the redirect
	// URL. Exchange will do the handshake to retrieve the
	// initial access token. The HTTP Client returned by
	// conf.Client will refresh the token as necessary.
	tok, err := conf.Exchange(ctx, <-code)
	if err != nil {
		log.Fatal(err)
	}

	client := conf.Client(ctx, tok)
	fmt.Println("OAuth access token: " + tok.AccessToken)
	fmt.Println("OAuth access token type: " + tok.TokenType)

	resp, err := client.Get("https://gitlab.com/api/v4/user")
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	prettyPrintJSON(body)
	done <- 0
}
