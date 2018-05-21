package main

import (
	"golang.org/x/oauth2/github"
)

func doGithubOAuthDance(done chan<- int, code <-chan string) {
	generateToken(
		github.Endpoint,
		[]string{"user"},
		"https://api.github.com/user",
		"https://github.com/settings/applications/new",
		done,
		code)
}
