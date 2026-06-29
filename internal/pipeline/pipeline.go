package pipeline

import (
	"context"

	"jobscout/internal/llm"
	"jobscout/internal/models"
	"jobscout/internal/scorer"
	"jobscout/internal/sources"
	"jobscout/internal/store"
	"jobscout/internal/yc"
)

type Result struct {
	Fetched  int `json:"fetched"`
	YCHiring int `json:"ycHiring"`
	Kept     int `json:"kept"`
	Added    int `json:"added"`
}

// Run executes one full scan: gather (incl. newly funded YC startups), score,
// classify, optionally refine with the LLM, and persist.
func Run(ctx context.Context, st *store.Store, c *llm.Client) (Result, error) {
	raw := sources.FetchAll(ctx)

	// Newly funded YC startups that are hiring + an index to label other jobs.
	ycIndex := map[string]string{}
	ycHiring := 0
	if cos, err := yc.LoadHiring(ctx); err == nil {
		raw = append(raw, sources.YCToRawJobs(cos)...)
		ycIndex = yc.Index(cos)
		ycHiring = len(cos)
	}

	jobs := make([]models.Job, 0, len(raw))
	for _, r := range raw {
		score, reason, elig := scorer.Keyword(r)
		// YC startup entries are kept even if keyword score is modest — the
		// whole point is to surface newly funded startups.
		if score < 20 && r.Source != "yc" {
			continue
		}
		j := r.ToModel(score, reason, elig)
		classify(&j, ycIndex)
		jobs = append(jobs, j)
	}

	// Refine the strongest candidates with Claude (if configured).
	candidates := make([]models.Job, 0)
	for _, j := range jobs {
		if j.Score >= 30 {
			candidates = append(candidates, j)
		}
	}
	if len(candidates) > 40 {
		candidates = candidates[:40]
	}
	verdicts := scorer.Refine(ctx, c, candidates)
	for i := range jobs {
		if v, ok := verdicts[jobs[i].ID]; ok {
			jobs[i].Score = v.Score
			jobs[i].Eligibility = v.Eligibility
			jobs[i].Reason = v.Reason
		}
	}

	added, err := st.UpsertJobs(jobs)
	if err != nil {
		return Result{}, err
	}
	return Result{Fetched: len(raw), YCHiring: ycHiring, Kept: len(jobs), Added: added}, nil
}
