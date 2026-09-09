package github

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRepositoryStateAndRulesets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case "/repos/example/project":
			writer.Write([]byte(`{"default_branch":"main","allow_squash_merge":true,"allow_merge_commit":false,"allow_rebase_merge":false,"delete_branch_on_merge":true}`))
		case "/repos/example/project/rulesets":
			writer.Write([]byte(`[{"id":1,"name":"main protection","target":"branch","enforcement":"active"}]`))
		case "/repos/example/project/labels":
			writer.Write([]byte(`[]`))
		default:
			t.Fatalf("unexpected path %s", request.URL.Path)
		}
		if request.URL.Path != "/repos/example/project" && request.Header.Get("Authorization") == "" {
			t.Fatalf("missing authorization")
		}
	}))
	defer server.Close()
	client := NewAdapter("example/project", "token")
	client.BaseURL = server.URL + "/repos"
	client.HTTP = server.Client()
	state, err := client.RepositoryState()
	if err != nil {
		t.Fatalf("repository: %v", err)
	}
	if state.DefaultBranch != "main" || !state.SquashMerge || state.MergeCommits {
		t.Fatalf("payload mapping: %#v", state)
	}
	rules, err := client.Rulesets()
	if err != nil || len(rules) != 1 || rules[0].Name != "main protection" {
		t.Fatalf("rulesets: %#v %v", rules, err)
	}
}

func TestRepositoryErrorsAreExplicit(t *testing.T) {
	for _, status := range []int{http.StatusUnauthorized, http.StatusNotFound, http.StatusInternalServerError} {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(status)
		}))
		client := NewAdapter("example/project", "token")
		client.BaseURL = server.URL + "/repos"
		client.HTTP = server.Client()
		if _, err := client.RepositoryState(); err == nil {
			t.Fatalf("expected status %d error", status)
		}
		server.Close()
	}
}
