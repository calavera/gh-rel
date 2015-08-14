package main

import "github.com/calavera/gh-rel/github"

func addProject(nwo string) error {
	return github.AddProject(nwo)
}
