package github_pagination

import (
	"io"
	"net/http"

	"github.com/gofri/go-github-pagination/github_pagination/json_merger"
	"github.com/gofri/go-github-pagination/github_pagination/pagination_utils"
)

type GitHubPagination struct {
	Base http.RoundTripper
}

func NewGithubPagination(base http.RoundTripper) *GitHubPagination {
	if base == nil {
		base = http.DefaultTransport
	}
	return &GitHubPagination{
		Base: base,
	}
}

func NewGithubPaginationClient(base http.RoundTripper) *http.Client {
	return &http.Client{
		Transport: NewGithubPagination(base),
	}
}

func (g *GitHubPagination) RoundTrip(request *http.Request) (resp *http.Response, err error) {
	merger := json_merger.NewMerger()
	for {
		resp, err = g.Base.RoundTrip(request)
		if err != nil {
			return resp, err
		}

		// only paginate through successful requests.
		// having a non-successful request would drop all previous results.
		// TODO this is gonna have a config for strictness (i.e., whether to drop all previous results or not)
		if resp.StatusCode != http.StatusOK {
			return resp, nil
		}

		if err := merger.ReadNext(resp.Body); err != nil {
			return resp, err
		}

		request = pagination_utils.GetNextRequest(request, resp)
		if request == nil {
			break
		}
	}

	resp.Body = io.NopCloser(merger.Merged())
	return resp, nil
}
