package db

import (
	"fmt"
	"os"
	"path"

	"github.com/boltdb/bolt"
)

var (
	internal       *bolt.DB
	projectsBucket = []byte("projects")
	etagsBucket    = []byte("etags")
	releasesBucket = []byte("releases")
)

func Open(dbPath string) (err error) {
	if err := os.MkdirAll(path.Dir(dbPath), 0755); err != nil {
		return err
	}
	internal, err = bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return err
	}
	return createBuckets()
}

func Close() error {
	if internal != nil {
		return internal.Close()
	}
	return nil
}

func createBuckets() error {
	return internal.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(projectsBucket); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(etagsBucket); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(releasesBucket); err != nil {
			return err
		}
		return nil
	})
}

func AddProject(name, htmlURL string) error {
	return internal.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(projectsBucket)

		n := []byte(name)
		if b.Get(n) != nil {
			return fmt.Errorf("the project already exists: %s", name)
		}
		b.Put(n, []byte(htmlURL))
		return nil
	})
}

func ListProjects() (projects []Project, err error) {
	err = internal.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(projectsBucket).Cursor()
		etags := tx.Bucket(etagsBucket)
		releases := tx.Bucket(releasesBucket)

		for k, v := c.First(); k != nil; k, v = c.Next() {
			name := string(k)
			latestEtag, releasesEtag := getEtags(etags, name)
			latestRel, nextRcRel := getReleases(releases, name)

			p := Project{
				Name:              name,
				HTMLURL:           string(v),
				LatestReleaseEtag: latestEtag,
				ReleasesEtag:      releasesEtag,
				LatestRelease:     latestRel,
				NextPreRelease:    nextRcRel,
			}

			projects = append(projects, p)
		}

		return nil
	})
	return
}

func SaveLatest(project, tag, url, etag string) error {
	return saveProjectRelease(project, "latest", tag, url, etag)
}

func SaveNextRcRelease(project, tag, url, etag string) error {
	return saveProjectRelease(project, "rc", tag, url, etag)
}

func saveProjectRelease(project, from, tag, url, etag string) error {
	return internal.Update(func(tx *bolt.Tx) error {
		if err := saveEtag(tx, project, from, etag); err != nil {
			return err
		}
		if err := saveRelease(tx, project, from, tag, url); err != nil {
			return err
		}
		return nil
	})
}

func getEtags(etags *bolt.Bucket, name string) (string, string) {
	latestEtagKey := fmt.Sprintf("%s-latest", name)
	releasesEtagKey := fmt.Sprintf("%s-releases-list", name)

	var latestEtag string
	if e := etags.Get([]byte(latestEtagKey)); e != nil {
		latestEtag = string(e)
	}
	var releasesEtag string
	if e := etags.Get([]byte(releasesEtagKey)); e != nil {
		releasesEtag = string(e)
	}
	return latestEtag, releasesEtag
}

func getReleases(releases *bolt.Bucket, name string) (Release, Release) {
	latestTagKey := fmt.Sprintf("%s-latest-tag", name)
	latestURLKey := fmt.Sprintf("%s-latest-url", name)

	rcTagKey := fmt.Sprintf("%s-rc-tag", name)
	rcURLKey := fmt.Sprintf("%s-rc-url", name)

	latest := wipRelease()
	rc := wipRelease()

	if t := releases.Get([]byte(latestTagKey)); t != nil {
		var u string
		if ub := releases.Get([]byte(latestURLKey)); ub != nil {
			u = string(ub)
		}
		latest = Release{string(t), u}
	}

	if t := releases.Get([]byte(rcTagKey)); t != nil {
		var u string
		if ub := releases.Get([]byte(rcURLKey)); ub != nil {
			u = string(ub)
		}
		rc = Release{string(t), u}
	}

	return latest, rc
}

func saveRelease(tx *bolt.Tx, name, from, tag, url string) error {
	b := tx.Bucket(releasesBucket)
	tagKey := fmt.Sprintf("%s-%s-tag", name, from)
	urlKey := fmt.Sprintf("%s-%s-url", name, from)

	if err := b.Put([]byte(tagKey), []byte(tag)); err != nil {
		return err
	}
	if err := b.Put([]byte(urlKey), []byte(url)); err != nil {
		return err
	}
	return nil
}

func saveEtag(tx *bolt.Tx, name, from, value string) error {
	b := tx.Bucket(etagsBucket)

	k := fmt.Sprintf("%s-%s", name, from)
	return b.Put([]byte(k), []byte(value))
}
