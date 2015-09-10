package db

import "strings"

const WIP = "WIP"

type Project struct {
	Name              string
	HTMLURL           string
	LatestReleaseEtag string
	ReleasesEtag      string
	LatestRelease     Release
	NextPreRelease    Release
}

func (p Project) Owner() string {
	return strings.Split(p.Name, "/")[0]
}

func (p Project) Repo() string {
	return strings.Split(p.Name, "/")[1]
}

type Release struct {
	Tag        string
	HTMLURL    string
	Prerelease bool
}

func wipRelease() Release {
	return Release{Tag: WIP}
}
