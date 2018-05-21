package main

import (
	"golang.org/x/oauth2/gitlab"
)

func doGitlabOAuthDance(done chan<- int, code <-chan string) {
	generateToken(
		gitlab.Endpoint,
		[]string{"api"},
		"https://gitlab.com/api/v4/user",
		"https://gitlab.com/profile/applications",
		done,
		code)
}
