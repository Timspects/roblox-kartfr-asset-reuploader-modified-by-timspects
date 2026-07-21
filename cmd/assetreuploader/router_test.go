package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Timspects/roblox-kartfr-asset-reuploader-modified-by-timspects/internal/app/response"
)

func TestWriteResponseProducesDoneForFinalState(t *testing.T) {
	rr := httptest.NewRecorder()
	resp := response.New()

	writeResponse(rr, resp, false, true, true)

	if rr.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusOK)
	}

	if contentType := rr.Header().Get("Content-Type"); !strings.Contains(contentType, "text/plain") {
		t.Fatalf("content type = %q, want text/plain", contentType)
	}

	if body := rr.Body.String(); body != "done" {
		t.Fatalf("body = %q, want done", body)
	}
}

func TestWriteResponseSendsPendingItemsBeforeDone(t *testing.T) {
	rr := httptest.NewRecorder()
	resp := response.New()
	resp.AddItem(response.ResponseItem{OldID: 101, NewID: 202})

	writeResponse(rr, resp, false, true, true)

	if rr.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusOK)
	}

	if contentType := rr.Header().Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		t.Fatalf("content type = %q, want application/json", contentType)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "101") || !strings.Contains(body, "202") {
		t.Fatalf("body = %q, want old/new IDs in JSON payload", body)
	}
}
