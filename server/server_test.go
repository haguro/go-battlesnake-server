package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/haguro/go-battlesnake-server/server"
)

func TestServer(t *testing.T) {
	info := server.InfoResponse{
		APIVersion: "1",
		Author:     "foo",
		Color:      "#000000",
		Head:       "default",
		Tail:       "default",
		Version:    "9.9",
	}
	moveResp := server.MoveResponse{
		Move:  "up",
		Shout: "Hi!",
	}
	logger := log.New(io.Discard, "", 0)
	s := server.New("0", &info, logger, 0, func(gs *server.GameState, l *server.Logger) server.MoveResponse { return moveResp })

	t.Run("IndexHandler", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		resp := httptest.NewRecorder()
		s.ServeHTTP(resp, req)

		if resp.Result().StatusCode != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, resp.Result().StatusCode)
		}
		if resp.Result().Header.Get("content-type") != "application/json" {
			t.Errorf("expected response with content-type 'application/json', got '%s'", resp.Result().Header.Get("content-type"))
		}

		var got server.InfoResponse
		decoder := json.NewDecoder(resp.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&got); err != nil {
			t.Fatalf("could not decode response body %q: %s", resp.Body.String(), err)
		}
		if got != info {
			t.Errorf("expected response body to be %q, got %q", info, got)
		}
	})

	t.Run("StartHandler", func(t *testing.T) {
		b := bytes.NewBuffer([]byte{})
		json.NewEncoder(b).Encode(&server.GameState{})
		req, _ := http.NewRequest(http.MethodPost, "/start", b)
		resp := httptest.NewRecorder()
		s.ServeHTTP(resp, req)

		if resp.Result().StatusCode != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, resp.Result().StatusCode)
		}
		if body := resp.Body.String(); body != "" {
			t.Errorf("expected response body to be empty, got %q", body)
		}
	})

	t.Run("StartHandlerBadRequest", func(t *testing.T) {
		b := bytes.NewBuffer([]byte("{invalid}"))
		req, _ := http.NewRequest(http.MethodPost, "/start", b)
		resp := httptest.NewRecorder()
		s.ServeHTTP(resp, req)

		if resp.Result().StatusCode != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d", http.StatusBadRequest, resp.Result().StatusCode)
		}
	})

	t.Run("EndHandler", func(t *testing.T) {
		b := bytes.NewBuffer([]byte{})
		json.NewEncoder(b).Encode(&server.GameState{})
		req, _ := http.NewRequest(http.MethodPost, "/end", b)
		resp := httptest.NewRecorder()
		s.ServeHTTP(resp, req)

		if resp.Result().StatusCode != http.StatusOK {
			t.Errorf("expected status code 200, got %v", resp.Result().StatusCode)
		}
		if body := resp.Body.String(); body != "" {
			t.Errorf("expected response body to be empty, got %q", body)
		}
	})

	t.Run("EndHandlerBadRequest", func(t *testing.T) {
		b := bytes.NewBuffer([]byte("{invalid}"))
		json.NewEncoder(b).Encode(&server.GameState{})
		req, _ := http.NewRequest(http.MethodPost, "/end", b)
		resp := httptest.NewRecorder()
		s.ServeHTTP(resp, req)

		if resp.Result().StatusCode != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d", http.StatusBadRequest, resp.Result().StatusCode)
		}
	})

	t.Run("MoveHandler", func(t *testing.T) {
		b := bytes.NewBuffer([]byte{})
		json.NewEncoder(b).Encode(&server.GameState{})
		req, _ := http.NewRequest(http.MethodPost, "/move", b)
		resp := httptest.NewRecorder()
		s.ServeHTTP(resp, req)

		if resp.Result().StatusCode != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, resp.Result().StatusCode)
		}
		if resp.Result().Header.Get("content-type") != "application/json" {
			t.Errorf("expected response with content-type 'application/json', got '%s'", resp.Result().Header.Get("content-type"))
		}

		var got server.MoveResponse
		decoder := json.NewDecoder(resp.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&got); err != nil {
			t.Fatalf("could not decode response body %q: %s", resp.Body.String(), err)
		}
		if got != moveResp {
			t.Errorf("expected response body to be %q, got %q", moveResp, got)
		}
	})

	t.Run("MoveHandlerBadRequest", func(t *testing.T) {
		b := bytes.NewBuffer([]byte("{invalid}"))
		json.NewEncoder(b).Encode(&server.GameState{})
		req, _ := http.NewRequest(http.MethodPost, "/move", b)
		resp := httptest.NewRecorder()
		s.ServeHTTP(resp, req)

		if resp.Result().StatusCode != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d", http.StatusBadRequest, resp.Result().StatusCode)
		}
	})

	t.Run("DebugRequestLogging", func(t *testing.T) {
		b := bytes.NewBuffer([]byte{})
		json.NewEncoder(b).Encode(&server.GameState{})
		req, _ := http.NewRequest(http.MethodPost, "/start", b)
		resp := httptest.NewRecorder()

		want := "DEBUG POST /start"
		logWriter := bytes.NewBuffer([]byte{})
		logger := log.New(logWriter, "", 0)
		s := server.New("0", &info, logger, server.LDebug, func(gs *server.GameState, l *server.Logger) server.MoveResponse { return moveResp })
		s.ServeHTTP(resp, req)
		logContent := logWriter.String()
		if !strings.Contains(logContent, want) {
			t.Errorf("expected log to contain %q", want)
		}

	})

	t.Run("InvalidURL", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/does-not-exist", nil)
		resp := httptest.NewRecorder()
		s.ServeHTTP(resp, req)

		if resp.Result().StatusCode != http.StatusNotFound {
			t.Errorf("expected status code 404, got %v", resp.Result().StatusCode)
		}
	})
}

func TestLogger(t *testing.T) {
	msg := "test log message"
	b := bytes.NewBuffer([]byte{})
	logger := server.NewLogger(log.New(b, "", log.LstdFlags), server.LError|server.LWarning|server.LInfo|server.LDebug)
	testCases := []struct {
		name      string
		f         func(string, ...any)
		expPrefix string
	}{
		{name: "InfoLogging", f: logger.Info, expPrefix: "INFO"},
		{name: "WarningLogging", f: logger.Warn, expPrefix: "WARN"},
		{name: "ErrorLogging", f: logger.Err, expPrefix: "ERROR"},
		{name: "DebugLogging", f: logger.Debug, expPrefix: "DEBUG"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			want := fmt.Sprintf("%s %s", tc.expPrefix, msg)
			tc.f(msg)
			if !strings.Contains(b.String(), want) {
				t.Errorf("expected log to contain %q", want)
			}
		})
	}
}
