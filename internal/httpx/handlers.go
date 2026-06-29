package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"jobscout/internal/apply"
	"jobscout/internal/config"
	"jobscout/internal/llm"
	"jobscout/internal/mail"
	"jobscout/internal/models"
	"jobscout/internal/pipeline"
	"jobscout/internal/store"
)

type Server struct {
	cfg config.Config
	st  *store.Store
	llm *llm.Client
}

func NewServer(cfg config.Config, st *store.Store, c *llm.Client) http.Handler {
	s := &Server{cfg: cfg, st: st, llm: c}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", s.health)
	mux.HandleFunc("GET /api/jobs", s.listJobs)
	mux.HandleFunc("PATCH /api/jobs/{id}", s.patchJob)
	mux.HandleFunc("POST /api/jobs/{id}/draft", s.draftApplication)
	mux.HandleFunc("POST /api/refresh", s.refresh)
	mux.HandleFunc("GET /api/applications", s.listApplications)
	mux.HandleFunc("PUT /api/applications/{id}", s.putApplication)
	mux.HandleFunc("POST /api/digest", s.digest)

	return logging(cors(s.cfg.AllowedOrigin, mux))
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"ok": true, "llm": s.llm.Enabled()})
}

func (s *Server) listJobs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	minScore, _ := strconv.Atoi(q.Get("minScore"))
	jobs := s.st.ListJobs(store.JobFilter{
		Status:      q.Get("status"),
		Source:      q.Get("source"),
		MinScore:    minScore,
		HideUSOnly:  q.Get("hideUsOnly") == "1",
		YC:          q.Get("yc") == "1",
		Funded:      q.Get("funded") == "1",
		NewlyFunded: q.Get("newlyFunded") == "1",
		Internship:  q.Get("internship") == "1",
	})
	writeJSON(w, 200, map[string]any{"jobs": jobs})
}

func (s *Server) patchJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Status models.TriageStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Status == "" {
		writeErr(w, 400, "status is required")
		return
	}
	job, err := s.st.SetJobStatus(id, body.Status)
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, 404, "job not found")
		return
	}
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"job": job})
}

func (s *Server) refresh(w http.ResponseWriter, r *http.Request) {
	res, err := pipeline.Run(r.Context(), s.st, s.llm)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, res)
}

// draftApplication generates tailored material and stores it as an Application.
func (s *Server) draftApplication(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	job, err := s.st.GetJob(id)
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, 404, "job not found")
		return
	}
	draft, err := apply.Generate(r.Context(), s.llm, job)
	if err != nil {
		writeErr(w, 422, err.Error())
		return
	}
	now := time.Now().UTC().Format(time.RFC3339)
	app := models.Application{
		JobID: id, Stage: models.StageDrafted,
		ResumeHighlights: draft.ResumeHighlights,
		CoverEmail:       draft.CoverEmail, Pitch: draft.Pitch,
		CreatedAt: now, UpdatedAt: now,
	}
	if existing, err := s.st.GetApplication(id); err == nil {
		app.CreatedAt = existing.CreatedAt
		app.Notes = existing.Notes
		if existing.Stage != "" && existing.Stage != models.StageDrafted {
			app.Stage = existing.Stage
		}
	}
	if err := s.st.SaveApplication(app); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"application": app})
}

func (s *Server) listApplications(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"applications": s.st.ListApplications()})
}

func (s *Server) putApplication(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Stage            models.AppStage `json:"stage"`
		Notes            string          `json:"notes"`
		ResumeHighlights string          `json:"resumeHighlights"`
		CoverEmail       string          `json:"coverEmail"`
		Pitch            string          `json:"pitch"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, "invalid body")
		return
	}
	app, err := s.st.GetApplication(id)
	if errors.Is(err, store.ErrNotFound) {
		app = models.Application{JobID: id, CreatedAt: time.Now().UTC().Format(time.RFC3339)}
	}
	if body.Stage != "" {
		app.Stage = body.Stage
	}
	if body.ResumeHighlights != "" {
		app.ResumeHighlights = body.ResumeHighlights
	}
	if body.CoverEmail != "" {
		app.CoverEmail = body.CoverEmail
	}
	if body.Pitch != "" {
		app.Pitch = body.Pitch
	}
	app.Notes = body.Notes
	app.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := s.st.SaveApplication(app); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"application": app})
}

func (s *Server) digest(w http.ResponseWriter, r *http.Request) {
	jobs := s.st.ListJobs(store.JobFilter{})
	sent, err := mail.SendDigest(r.Context(), s.cfg, jobs)
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"sent": sent})
}
