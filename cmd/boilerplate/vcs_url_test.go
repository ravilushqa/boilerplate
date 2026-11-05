package main

import (
	"net"
	"strings"
	"testing"
)

func TestParseVCSUrl(t *testing.T) {
	repos := []string{
		"ssh://git@github.com:7875/ravilushqa/boilerplate.git",
		"git://github.com:7875/ravilushqa/boilerplate.git",
		"https://github.com:7875/ravilushqa/boilerplate.git",
		"ftps://github.com:7875/ravilushqa/boilerplate.git",
		"git@github.com:ravilushqa/boilerplate.git",
		"ssh://git@github.com:7875/ravilushqa/boilerplate.git",
		"git://github.com:7875/ravilushqa/boilerplate.git",
		"git@github.com:ravilushqa/boilerplate.git",
		"~/ravilushqa/boilerplate.git",
		"file://~/ravilushqa/boilerplate.git",
	}
	for _, repo := range repos {
		url, err := ParseVCSUrl(repo)
		if err != nil {
			t.Fatal(repo, err)
		}
		urlPath := strings.TrimLeft(url.Path, "/")
		if urlPath != "ravilushqa/boilerplate.git" {
			t.Fatal(repo, "parse url failed", urlPath)
		}
	}
}

func TestParseSsh(t *testing.T) {
	repo := "ssh://git@github.com:7875/ravilushqa/boilerplate.git"
	url, err := ParseVCSUrl(repo)
	if err != nil {
		t.Fatal(err)
	}
	host, _, err := net.SplitHostPort(url.Host)
	if err != nil {
		host = url.Host
	}
	t.Log(host, url.Path)
}
