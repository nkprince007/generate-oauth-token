# generate-oauth-token

A CLI tool to generate OAuth tokens for the chosen provider.

## Installation

Download the tool using `go get`.

```bash
go get -u -v github.com/nkprince007/generate-oauth-token
```

## Usage

```bash
$ generate-oauth-token -help
Usage of generate-oauth-token:
  -provider string
      the provider for which oauth token is to be generated (default "github")
```

```bash
$ generate-oauth-token -provider=github
...
# Enter the details and proceed to obtain the token
```
