package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"time"
)

var organizations = []string{
	"vshn",
	"appuio",
	"projectsyn",
	"k8up-io",
}

var (
	GithubUsername = os.Getenv("GITHUB_USERNAME")
	GithubPassword = os.Getenv("GITHUB_PASSWORD")
)

type Repo struct {
	FullName    string `json:"full_name"`
	Description string `json:"description"`

	Fork bool `json:"fork"`

	UpdatedAt time.Time `json:"updated_at"`
	PushedAt  time.Time `json:"pushed_at"`
}

func (r Repo) String() string {
	forkStr := ""
	if r.Fork {
		forkStr = "FORKED "
	}
	return fmt.Sprintf("%s%s (%s) [%s]", forkStr, r.FullName, r.Description, r.LastUpdated())
}

func (r Repo) LastUpdated() time.Time {
	if r.UpdatedAt.After(r.PushedAt) {
		return r.UpdatedAt
	}
	return r.PushedAt
}

func main() {
	for _, organization := range organizations {
		fetchForOrg(organization)
	}
}

func fetchForOrg(org string) {
	fetched := make([]Repo, 0)
	client := &http.Client{}

	page := 1
	for {
		req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/orgs/%s/repos?per_page=100&page=%d", org, page), nil)
		if err != nil {
			panic(err)
		}
		if GithubUsername != "" {
			req.SetBasicAuth(GithubUsername, GithubPassword)
		}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		repos := make([]Repo, 0, 100)
		err = json.NewDecoder(resp.Body).Decode(&repos)
		if err != nil {
			panic(err)
		}

		fetched = append(fetched, repos...)
		page += 1

		if len(repos) == 0 {
			break
		}
	}

	fmt.Fprintf(os.Stderr, "Got %d repos for org %s ...\n", len(fetched), org)

	sort.Slice(fetched, func(i, j int) bool {
		return fetched[i].LastUpdated().After(fetched[j].LastUpdated())
	})

	for _, r := range fetched {
		fmt.Println("- ", r)
	}
	fmt.Print("\n")
}
