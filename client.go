package main

import (
	"strings"

	"github.com/calavera/gh-rel/db"
	"github.com/calavera/gh-rel/github"
)

const defaultOrg = "docker"

func addProject(nwo string) error {
	nameWithOwner := strings.Split(nwo, "/")
	if len(nameWithOwner) == 1 {
		nameWithOwner = []string{defaultOrg, nameWithOwner[0]}
	}

	p, r := github.Project(nameWithOwner[0], nameWithOwner[1])
	if r.HasError() {
		return r.Err
	}

	return db.AddProject(p.FullName, p.HTMLURL)
}
