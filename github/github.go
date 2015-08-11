package github

import (
	"net/url"

	"github.com/octokit/go-octokit/octokit"
)

const (
	etagKey      = "ETag"
	noneMatchKey = "If-None-Match"
)

var internal *octokit.Client

func InitClient(token string) {
	var auth octokit.AuthMethod
	if token != "" {
		auth = octokit.TokenAuth{token}
	}
	internal = octokit.NewClientWith("https://api.github.com", "gh-rel-dashboard", auth, nil)
}

func Project(owner, name string) (*octokit.Repository, *octokit.Result) {
	return internal.Repositories().One(nil, repoParams(owner, name))
}

func LatestRelease(owner, name, etag string) (release *octokit.Release, result *octokit.Result, respEtag string) {
	result, respEtag = releaseRequest(owner, name, etag, octokit.ReleasesLatestURL, &release)
	return
}

func NextRcRelease(owner, name, etag string) (*octokit.Release, *octokit.Result, string) {
	var releases []octokit.Release
	result, respEtag := releaseRequest(owner, name, etag, octokit.ReleasesURL, &releases)
	if result.HasError() {
		return nil, result, ""
	}

	if len(releases) == 0 {
		return nil, result, ""
	}

	for _, r := range releases {
		if !r.Draft && !r.Prerelease {
			break
		}
		if r.Prerelease {
			return &r, result, respEtag
		}
	}

	return nil, result, ""
}

func releaseRequest(owner, name, etag string, uri octokit.Hyperlink, output interface{}) (result *octokit.Result, respEtag string) {
	params := repoParams(owner, name)
	url, err := uri.Expand(params)
	if err != nil {
		return &octokit.Result{Err: err}, ""
	}

	result, respEtag = sendRequest(url, func(req *octokit.Request) (*octokit.Response, error) {
		req.Header.Set(noneMatchKey, etag)
		return req.Get(output)
	})

	return
}

func repoParams(owner, name string) octokit.M {
	return octokit.M{"owner": owner, "repo": name}
}

func sendRequest(url *url.URL, fn func(r *octokit.Request) (*octokit.Response, error)) (result *octokit.Result, respEtag string) {
	req, err := internal.NewRequest(url.String())
	if err != nil {
		result = &octokit.Result{Response: nil, Err: err}
		return
	}

	resp, err := fn(req)
	result = &octokit.Result{Response: resp, Err: err}
	if resp != nil {
		respEtag = resp.Header.Get(etagKey)
	}

	return
}
