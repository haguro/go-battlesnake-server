package server

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func BenchmarkMoveHandler(b *testing.B) {
	info := InfoResponse{
		APIVersion: "1",
		Author:     "foo",
		Color:      "#000000",
		Head:       "default",
		Tail:       "default",
		Version:    "9.9",
	}

	logger := log.New(io.Discard, "", 0)
	s := New("0", &info, logger, 0, func(gs *GameState, l *Logger) MoveResponse {
		return MoveResponse{
			Move:  "up",
			Shout: "Hi!",
		}
	})

	requestBody, err := os.ReadFile("move_test.json")
	if err != nil {
		b.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/move", bytes.NewBuffer(requestBody))
	if err != nil {
		b.Fatal(err)
	}
	rr := httptest.NewRecorder()

	handler := s.moveHandler()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler(rr, req)
		if status := rr.Code; status != http.StatusOK {
			b.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
	}
}
