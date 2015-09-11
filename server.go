package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/calavera/gh-rel/db"
	"github.com/calavera/gh-rel/github"
	"github.com/calavera/gh-rel/render"
	"github.com/gin-gonic/gin"
)

type release struct {
	Tag        string
	URL        string
	prerelease bool
}

func (r release) WIP() bool {
	return r.Tag == db.WIP
}

func (r release) Label() string {
	if r.WIP() {
		return "wip"
	}
	if r.prerelease || strings.Contains(strings.ToLower(r.Tag), "rc") {
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

type add struct {
	Error error
}

func startServer(port uint, adminPassword string) {
	router := gin.Default()
	router.Static("/assets", "./assets")

	router.GET("/", func(c *gin.Context) {
		showProjects(c)
	})

	authorized := router.Group("/add", gin.BasicAuth(gin.Accounts{
		"admin": adminPassword,
	}))

	authorized.GET("", func(c *gin.Context) {
		render.New(c).HTML(http.StatusOK, "add.tmpl", nil)
	})

	authorized.POST("", func(c *gin.Context) {
		addNewProject(c)
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

	render.New(c).HTML(http.StatusOK, "index.tmpl", index{prs})
}

func addNewProject(c *gin.Context) {
	nwo := strings.TrimSpace(c.PostForm("repo"))
	if nwo == "" {
		render.New(c).HTML(http.StatusOK, "add.tmpl", add{fmt.Errorf("the repository name cannot be empty")})
		return
	}

	if err := github.AddProject(nwo); err != nil {
		render.New(c).HTML(http.StatusOK, "add.tmpl", add{err})
		return
	}

	showProjects(c)
}

func getLatestRelease(p db.Project) release {
	rel, res, etag := github.LatestRelease(p.Owner(), p.Repo(), p.LatestReleaseEtag)
	if res.HasError() {
		log.Println(res.Err)
		return release{p.LatestRelease.Tag, p.LatestRelease.HTMLURL, false}
	}

	if res.Response.StatusCode == http.StatusOK {
		db.SaveLatest(p.Name, rel.TagName, rel.HTMLURL, etag)
		return release{rel.TagName, rel.HTMLURL, false}
	}
	return release{p.LatestRelease.Tag, p.LatestRelease.HTMLURL, false}
}

func getRcRelease(p db.Project) release {
	rel, res, etag := github.NextRcRelease(p.Owner(), p.Repo(), p.ReleasesEtag)
	if res.HasError() {
		log.Println(res.Err)
		return release{p.LatestRelease.Tag, p.LatestRelease.HTMLURL, true}
	}

	if rel == nil {
		return release{Tag: db.WIP}
	}

	if res.Response.StatusCode == http.StatusOK {
		db.SaveNextRcRelease(p.Name, rel.TagName, rel.HTMLURL, etag)
		return release{rel.TagName, rel.HTMLURL, true}
	}
	return release{p.NextPreRelease.Tag, p.NextPreRelease.HTMLURL, true}
}
