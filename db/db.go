package db

import (
	"encoding/json"
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
	return saveProjectRelease(project, "latest", tag, url, etag, false)
}

func SaveNextRcRelease(project, tag, url, etag string) error {
	return saveProjectRelease(project, "rc", tag, url, etag, true)
}

func saveProjectRelease(project, from, tag, url, etag string, prerelease bool) error {
	return internal.Update(func(tx *bolt.Tx) error {
		if err := saveEtag(tx, project, from, etag); err != nil {
			return err
		}
		if err := saveRelease(tx, project, from, tag, url, prerelease); err != nil {
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
	latestKey := fmt.Sprintf("%s-latest", name)
	rcKey := fmt.Sprintf("%s-rc", name)

	latest := wipRelease()
	rc := wipRelease()

	if t := releases.Get([]byte(latestKey)); t != nil {
		json.Unmarshal(t, &latest)
	}

	if t := releases.Get([]byte(rcKey)); t != nil {
		json.Unmarshal(t, &rc)
	}

	return latest, rc
}

func saveRelease(tx *bolt.Tx, name, from, tag, url string, prerelease bool) error {
	b := tx.Bucket(releasesBucket)
	key := fmt.Sprintf("%s-%s", name, from)

	release := Release{tag, url, prerelease}
	bytes, err := json.Marshal(release)
	if err != nil {
		return err
	}

	return b.Put([]byte(key), bytes)
}

func saveEtag(tx *bolt.Tx, name, from, value string) error {
	b := tx.Bucket(etagsBucket)

	k := fmt.Sprintf("%s-%s", name, from)
	return b.Put([]byte(k), []byte(value))
}
