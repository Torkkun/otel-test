package server

import (
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
)

func handlerSingle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sleepTime := randomSleep(r)
		fmt.Fprintf(w, "work completed in %v\n", sleepTime)
	}
}

func handlerMulti() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		subRequests := 3 + rand.Intn(4)
		// Write a structured log with the request context, which allows the log to
		// be linked with the trace for this request.
		slog.InfoContext(r.Context(), "handle /multi request", slog.Int("subRequests", subRequests))

		err := computeSubrequests(r, subRequests)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		fmt.Fprintln(w, "ok")
	}
}
