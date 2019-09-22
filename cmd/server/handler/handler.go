package handler

import (
	"math/rand"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type loginHandler struct {
	logger       *log.Logger
	correctEmail string
	baseLatency  float64
	stdDev       float64
}

// NewNaiveComparator returns a handler which does short-circuit rune comapration
func NewNaiveComparator(l *log.Logger, baseLatency, stdDev int) http.Handler {
	return &loginHandler{
		logger:       l,
		correctEmail: "correct@email.com",
		baseLatency:  float64(baseLatency),
		stdDev:       float64(stdDev),
	}
}

func (h loginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")

	noise := time.Duration(h.baseLatency + rand.NormFloat64()*h.stdDev)
	if noise < 0 {
		noise = 0
	}
	time.Sleep(noise * time.Millisecond)

	if email != h.correctEmail {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}
	time.Sleep(500 * time.Microsecond) // Simulate failing password check
	http.Error(w, "", http.StatusUnauthorized)
	return
}
