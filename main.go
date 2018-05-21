package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
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

func main() {
	var provider string

	flag.StringVar(
		&provider,
		"provider",
		"github",
		"the provider for which oauth token is to be generated")
	flag.Parse()

	fmt.Println("Use OAuth application server homepage as " +
		"'http://localhost:8000/'")
	done := make(chan int)

	switch provider {
	case "github":
		go doGithubOAuthDance(done, code)
	default:
		fmt.Println("generate-oauth-token: Incorrect usage")
		flag.Usage()
		os.Exit(1)
	}

	startServer(done)
}
