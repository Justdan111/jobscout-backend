package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"jobscout/internal/models"
)

var ErrNotFound = errors.New("not found")

// Store persists jobs and applications to JSON files. It is safe for
// concurrent use. The interface is small on purpose: to move to Postgres,
// reimplement these methods and leave the rest of the app untouched.
type Store struct {
	mu   sync.RWMutex
	dir  string
	jobs map[string]models.Job
	apps map[string]models.Application
}

func New(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	s := &Store{dir: dir, jobs: map[string]models.Job{}, apps: map[string]models.Application{}}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) load() error {
	if err := readJSON(filepath.Join(s.dir, "jobs.json"), &s.jobs); err != nil {
		return err
	}
	if err := readJSON(filepath.Join(s.dir, "applications.json"), &s.apps); err != nil {
		return err
	}
	if len(s.jobs) == 0 {
		for _, j := range seedJobs() {
			s.jobs[j.ID] = j
		}
		_ = writeJSON(filepath.Join(s.dir, "jobs.json"), s.jobs)
	}
	return nil
}

// ---- Jobs ----

type JobFilter struct {
	Status      string
	Source      string
	MinScore    int
	HideUSOnly  bool
	YC          bool
	Funded      bool
	NewlyFunded bool
	Internship  bool
}

func (s *Store) ListJobs(f JobFilter) []models.Job {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Job, 0, len(s.jobs))
	for _, j := range s.jobs {
		if f.Status != "" && f.Status != "all" && string(j.Status) != f.Status {
			continue
		}
		if f.Source != "" && f.Source != "all" && j.Source != f.Source {
			continue
		}
		if j.Score < f.MinScore {
			continue
		}
		if f.HideUSOnly && j.Eligibility == models.EligUSOnly {
			continue
		}
		if f.YC && !j.YC {
			continue
		}
		if f.Funded && !j.Funded {
			continue
		}
		if f.NewlyFunded && !j.NewlyFunded {
			continue
		}
		if f.Internship && !j.Internship {
			continue
		}
		out = append(out, j)
	}
	sort.Slice(out, func(i, k int) bool { return out[i].Score > out[k].Score })
	return out
}

func (s *Store) GetJob(id string) (models.Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	j, ok := s.jobs[id]
	if !ok {
		return models.Job{}, ErrNotFound
	}
	return j, nil
}

// UpsertJobs adds new jobs, preserving the triage status of existing ones.
func (s *Store) UpsertJobs(jobs []models.Job) (added int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, j := range jobs {
		if prev, ok := s.jobs[j.ID]; ok {
			j.Status = prev.Status
			j.DiscoveredAt = prev.DiscoveredAt
		} else {
			added++
		}
		s.jobs[j.ID] = j
	}
	return added, writeJSON(filepath.Join(s.dir, "jobs.json"), s.jobs)
}

func (s *Store) SetJobStatus(id string, status models.TriageStatus) (models.Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, ok := s.jobs[id]
	if !ok {
		return models.Job{}, ErrNotFound
	}
	j.Status = status
	s.jobs[id] = j
	return j, writeJSON(filepath.Join(s.dir, "jobs.json"), s.jobs)
}

// ---- Applications ----

func (s *Store) ListApplications() []models.Application {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Application, 0, len(s.apps))
	for _, a := range s.apps {
		out = append(out, a)
	}
	sort.Slice(out, func(i, k int) bool { return out[i].UpdatedAt > out[k].UpdatedAt })
	return out
}

func (s *Store) GetApplication(jobID string) (models.Application, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.apps[jobID]
	if !ok {
		return models.Application{}, ErrNotFound
	}
	return a, nil
}

func (s *Store) SaveApplication(a models.Application) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.apps[a.JobID] = a
	return writeJSON(filepath.Join(s.dir, "applications.json"), s.apps)
}

// ---- helpers ----

func readJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, v)
}

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
