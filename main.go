package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	"golang.org/x/oauth2"
)

// Provider indicates the hosting provider
type Provider string

const (
	// GitHub indicates that the required provider is github
	GitHub Provider = "github"
	// GitLab indicates that the required provider is gitlab
	GitLab Provider = "gitlab"
)

var code = make(chan string, 1)

func prettyPrintJSON(body []byte) {
	var prettyJSON bytes.Buffer
	error := json.Indent(&prettyJSON, body, "", "\t")
	if error != nil {
		log.Fatal("JSON parse error: ", error)
		return
	}
	fmt.Println(string(prettyJSON.Bytes()))
}

func readSecretFromStdin(prompt string) string {
	fmt.Println(prompt)
	byteSecret, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal(err)
	}
	return string(byteSecret)
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	case "windows":
		err = exec.Command(
			"rundll32", "url.dll,FileProtocolHandler", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func extractCode(w http.ResponseWriter, r *http.Request) {
	codes, ok := r.URL.Query()["code"]
	if !ok || len(codes) < 1 {
		io.WriteString(w, "Url Param 'code' missing")
	} else {
		c := codes[0]
		io.WriteString(w, "Code: "+c)
		go func() { code <- c }()
	}
}

func startServer(done chan int) {
	errs := make(chan error)
	go func() {
		c := make(chan os.Signal, 1)
		defer close(c)

		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("SIGNAL (%s)", <-c)
	}()

	http.HandleFunc("/", extractCode)
	go func() {
		errs <- http.ListenAndServe(":8000", nil)
	}()

	select {
	case status := <-done:
		os.Exit(status)
	case err := <-errs:
		fmt.Printf("\nReceived: %v\n", err)
	}
}

func generateToken(
	endpoint oauth2.Endpoint, scopes []string, testURL string, newAppURL string,
	done chan<- int, code <-chan string,
) {
	ctx := context.Background()
	fmt.Println("Create your OAuth application here (" + newAppURL + ")")
	clientID := readSecretFromStdin("Enter your client ID please: ")
	clientSecret := readSecretFromStdin("Enter your client secret please: ")

	fmt.Print("Enter your OAuth scopes please (seperated by space, enter EOF " +
		"or ^D to stop): ")
	for {
		var scope string
		if _, err := fmt.Scan(&scope); err != nil {
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
		Endpoint:     endpoint,
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

	if byteToken, err := json.Marshal(*tok); err != nil {
		log.Fatal(err)
	} else {
		prettyPrintJSON(byteToken)
	}

	client := conf.Client(ctx, tok)
	resp, err := client.Get(testURL)
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

func main() {
	var provider string

	flag.StringVar(
		&provider,
		"provider",
		"github",
		"the provider for which oauth token is to be generated, one of "+
			"('github', 'gitlab')")
	flag.Parse()

	fmt.Println("Use OAuth application server homepage as " +
		"'http://localhost:8000/'")
	done := make(chan int)

	switch Provider(provider) {
	case GitHub:
		go doGithubOAuthDance(done, code)
	case GitLab:
		go doGitlabOAuthDance(done, code)
	default:
		fmt.Println("generate-oauth-token: Incorrect usage")
		flag.Usage()
		os.Exit(1)
	}

	startServer(done)
}
