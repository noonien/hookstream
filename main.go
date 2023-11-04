package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var (
	addr   = flag.String("addr", ":8080", "address to listen on")
	prefix = flag.String("prefix", "/", "http prefix to serve on")
)

type Message struct {
	Topic   string         `json:"topic"`
	Method  string         `json:"method"`
	Headers map[string]any `json:"headers"`
	Query   string         `json:"query,omitempty"`
	Data    any            `json:"data,omitempty"`
}

func main() {
	flag.Parse()

	// allow all cross-origin requests
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	prefix := "/" + strings.TrimPrefix(*prefix, "/")

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Route(prefix, func(r chi.Router) {
		r.HandleFunc("/hook/{topic}", handleHook)
		r.Get("/socket", handleSocket)
		r.Get("/socket/{topic}", handleSocket)
	})

	log.Println("listening on", *addr)
	err := http.ListenAndServe(*addr, r)
	if err != nil {
		log.Fatal(err)
	}
}

// handleHook processes incoming webhooks and broadcasts their data and metadata to subscribers
func handleHook(w http.ResponseWriter, r *http.Request) {
	topic := chi.URLParam(r, "topic")

	headers := make(map[string]any, len(r.Header))
	for name, values := range r.Header {
		headers[name] = values[0]
		if len(values) > 1 {
			headers[name] = values
		}
	}

	msg := Message{
		Topic:   topic,
		Method:  r.Method,
		Query:   r.URL.RawQuery,
		Headers: headers,
	}

	body, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		writeError(w, r, http.StatusBadRequest, err)
		return
	}

	// include JSON as-is
	if json.Valid(body) {
		msg.Data = json.RawMessage(body)
	} else {
		msg.Data = string(body)
	}

	data, err := json.Marshal(msg)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusOK)

	go pub(topic, data)
}

func writeError(w http.ResponseWriter, r *http.Request, code int, err error) {
	var resp struct {
		Error string `json:"error"`
	}
	resp.Error = err.Error()

	w.WriteHeader(code)
	json.NewEncoder(w).Encode(resp)
	return
}
