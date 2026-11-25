package api

import (
	"context"
	"encoding/json"
	"errors"
	"llm_gateway/internal/auth"
	"llm_gateway/internal/cache"
	"llm_gateway/internal/events"
	"llm_gateway/internal/job"
	"llm_gateway/internal/limiter"
	"llm_gateway/internal/router"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Handler struct {
	auth      *auth.Auth
	limiter   *limiter.Limiter
	jobSvc    *job.Service
	router    *router.Router
	cache     *cache.Cache
	events    events.Publisher
	logger    *slog.Logger
	depChecks []depCheck
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) routes() http.Handler {
	mux := http.NewServeMux()

	protected := func(hf http.HandlerFunc) http.Handler {
		return h.auth.Middleware(h.limiter.Middleware(hf))
	}

	mux.Handle("POST /v1/job", protected(h.handleSubmitJob))
	mux.Handle("POST /v1/chat", protected(h.handleChat))
	mux.Handle("GET /v1/job/{id}", protected(h.handleGetJob))
	mux.HandleFunc("GET /readyz", h.handleReady)
	mux.HandleFunc("GET /healthz", h.handleHealth)
	return mux
}

func (h *Handler) handleSubmitJob(w http.ResponseWriter, r *http.Request) {
	caller, _ := auth.GetCallerFromContext(r.Context())

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		writeError(w, http.StatusBadRequest, "Idempotency-Key header required")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var p job.RequestPayload

	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if err := p.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	j, err := h.jobSvc.SubmitJob(r.Context(), caller.ID, idemKey, p)
	if err != nil {
		if errors.Is(err, job.ErrKeyReuse) {
			writeError(w, http.StatusConflict, "idempotency key reused with a different request")
			return
		}
		slog.Error("submit failed", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJson(w, http.StatusAccepted, map[string]any{"job_id": j.ID, "job_state": j.State})
}

func (h *Handler) handleGetJob(w http.ResponseWriter, r *http.Request) {
	caller, _ := auth.GetCallerFromContext(r.Context())
	id, err := uuid.Parse(r.PathValue("id"))

	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid job id")
	}

	j, err := h.jobSvc.GetJob(r.Context(), caller.ID, id)
	if err != nil {
		if errors.Is(err, job.ErrNotFound) {
			writeError(w, http.StatusNotFound, "job not found")
			return
		}
		slog.Error("get job failed", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJson(w, http.StatusOK, j)

}

func (h *Handler) handleChat(w http.ResponseWriter, r *http.Request) {
	caller, ok := auth.GetCallerFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var p job.RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if err := p.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	hash, _ := p.Hash()
	key := "resp:" + caller.ID.String() + ":" + hash

	var resp job.ResponsePayload
	if hit, _ := h.cache.Get(r.Context(), key, resp); hit {

	}

	out, err := h.router.Complete(r.Context(), p)
	if err != nil {
		slog.Error("all providers failed", "err", err)
		return
	}

	_ = h.cache.Set(r.Context(), key, out)
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (h *Handler) handleReady(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	var mu sync.Mutex
	var wg sync.WaitGroup

	statusChecks := make(map[string]string, len(h.depChecks))
	ready := true

	for _, c := range h.depChecks {
		wg.Add(1)
		go func(c depCheck) {
			defer wg.Done()
			err := c.Ping(ctx)
			mu.Lock()
			if err != nil {
				statusChecks[c.Name] = "down"
				if c.Critical {
					ready = false
				}
			} else {
				statusChecks[c.Name] = "up"
			}
			mu.Unlock()
		}(c)
	}

	wg.Wait()

	status := http.StatusOK
	if !ready {
		status = http.StatusServiceUnavailable
	}

	writeJson(w, status, map[string]any{"ready": ready, "dep_checks": statusChecks})
}
