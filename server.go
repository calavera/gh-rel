package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/calavera/gh-rel/db"
	"github.com/calavera/gh-rel/github"
	"github.com/gin-gonic/gin"
)

type release struct {
	Tag string
	URL string
}

func (r release) WIP() bool {
	return r.Tag == db.WIP
}

func (r release) Label() string {
	if r.WIP() {
		return "wip"
	}
	if strings.Contains(strings.ToLower(r.Tag), "rc") {
		return "prerelease"
	}
	return "latest"
}

type project struct {
	Owner          string
	Repo           string
	LatestRelease  release
	NextPreRelease release
}

type index struct {
	Projects []project
}

func startServer(port uint) {
	router := gin.Default()
	router.Static("/assets", "./assets")
	router.LoadHTMLGlob("templates/*")

	router.GET("/", func(c *gin.Context) {
		showProjects(c)
	})

	router.Run(fmt.Sprintf(":%v", port))
}

func showProjects(c *gin.Context) {
	projects, err := db.ListProjects()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.tmpl", nil)
		return
	}

	var prs []project
	for _, p := range projects {
		lr := getLatestRelease(p)
		rc := getRcRelease(p)

		prs = append(prs, project{p.Owner(), p.Repo(), lr, rc})
	}

	c.HTML(http.StatusOK, "index.tmpl", index{prs})
}

func getLatestRelease(p db.Project) release {
	rel, res, etag := github.LatestRelease(p.Owner(), p.Repo(), p.LatestReleaseEtag)
	if res.HasError() {
		log.Println(res.Err)
		return release{p.LatestRelease.Tag, p.LatestRelease.HTMLURL}
	}

	if res.Response.StatusCode == http.StatusOK {
		db.SaveLatest(p.Name, rel.TagName, rel.HTMLURL, etag)
		return release{rel.TagName, rel.HTMLURL}
	}
	return release{p.LatestRelease.Tag, p.LatestRelease.HTMLURL}
}

func getRcRelease(p db.Project) release {
	rel, res, etag := github.NextRcRelease(p.Owner(), p.Repo(), p.ReleasesEtag)
	if res.HasError() {
		log.Println(res.Err)
		return release{p.LatestRelease.Tag, p.LatestRelease.HTMLURL}
	}

	if rel == nil {
		return release{Tag: db.WIP}
	}

	if res.Response.StatusCode == http.StatusOK {
		db.SaveNextRcRelease(p.Name, rel.TagName, rel.HTMLURL, etag)
		return release{rel.TagName, rel.HTMLURL}
	}
	return release{p.NextPreRelease.Tag, p.NextPreRelease.HTMLURL}
}
