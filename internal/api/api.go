package api

import (
	"context"
	"llm_gateway/internal/auth"
	"llm_gateway/internal/cache"
	"llm_gateway/internal/events"
	"llm_gateway/internal/router"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

type Handler struct {
	auth    *auth.Auth
	limiter *ratelimit.Limiter
	jobSvc  *jobs.Service
	router  *router.Router
	cache   *cache.Cache
	events  events.Publisher
	logger  *slog.Logger
}

func NewHandler() *Handler {
	return &Handler{}
}

type depChecks 

type depCheck struct {
	name     string
	critical bool
	ping     func(ctx context.Context) error
}

func (h *Handler) routes() http.Handler {
	mux := http.NewServeMux()

	protected := func(hf http.HandlerFunc) http.Handler {
		return h.auth.Middleware(h.limiter.Middleware(hf))
	}
	mux.HandleFunc("POST /v1/job", protected(h.handleSubmitJob))
	mux.HandleFunc("POST /v1/chat", protected(h.handleChat))
	mux.HandleFunc("GET /v1/job/{id}", protected(h.handleGetJob))
	mux.HandleFunc("GET /healthz", h.handleReady)
	mux.HandleFunc("GET /readyz", h.HandleReady)
	return mux
}

func (h *Handler) handleSubmitJob(w http.ResponseWriter, r *http.Request) {
	caller, _ := auth.GetCallerFromContext(r.Context())
	idemKey := r.Header.Get("Idempotency-Key")

	writeJson(w, http.StatusAccepted)
}

func (h *Handler) handleGetJob(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) handleChat(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriterHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (h *Handler) handleReady(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	var mu sync.Mutex
	var wg sync.WaitGroup

	statusChecks := make(map[string]string)
	ready := true

	for _, c := range h.depChecks {
		wg.Add(1)
		go func(c depCheck) {
			defer wg.Done()
			err := c.ping(ctx)
			mu.Lock()
			if err != nil {
				results[c.name] = "down"
				if c.critical {
					ready = false
				}
			} else {
				results[c.name] = "up"
			}
			mu.Unlock()
		}(c)
	}

	wg.Wait()

	status := http.StatusOK
	if !ready {
		status = http.StatusServiceUnavailable
	}

	writeJson()
}
