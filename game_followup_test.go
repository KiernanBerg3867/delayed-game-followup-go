package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestScheduleFollowUpOnlyForPublishedMatchingAsset(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.Write([]byte(`{"ok":true,"data":{},"error":null,"metadata":{}}`))
	}))
	defer server.Close()
	client := &Client{key: "test", httpClient: server.Client(), baseURL: server.URL}
	err := scheduleFollowUp(client, Asset{PlayerID: "p", AssetID: "a", Published: false}, LiveEvent{AssetID: "a", FollowUp: "reward"})
	if err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("unpublished assets must not create a delayed action")
	}
	err = scheduleFollowUp(client, Asset{PlayerID: "p", AssetID: "a", Published: true}, LiveEvent{AssetID: "a", FollowUp: "reward"})
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("published matching assets must queue a delayed action")
	}
}
