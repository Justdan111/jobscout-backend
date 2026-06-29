package scorer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"jobscout/internal/llm"
	"jobscout/internal/models"
	"jobscout/internal/profile"
	"jobscout/internal/sources"
)

// Keyword scores a raw posting cheaply (no API). Returns 0-100, a reason, and
// the location eligibility guess.
func Keyword(r sources.RawJob) (int, string, models.Eligibility) {
	title := strings.ToLower(r.Title)
	blob := strings.ToLower(strings.Join([]string{
		r.Title, r.Company, r.Location, strings.Join(r.Tags, " "), r.Description,
	}, " "))

	score := 0
	var why []string

	if hit := firstHit(title, profile.Roles); hit != "" {
		score += 35
		why = append(why, "role: "+hit)
	}
	if hits := allHits(blob, profile.Stack); len(hits) > 0 {
		score += min(30, len(hits)*12)
		why = append(why, "stack: "+strings.Join(firstN(hits, 3), ", "))
	}
	if hit := firstHit(blob, profile.GoodSeniority); hit != "" {
		score += 15
		why = append(why, "level: "+hit)
	}
	if hit := firstHit(blob, profile.HardSeniority); hit != "" {
		score -= 20
		why = append(why, "stretch: "+hit)
	}

	elig := detectEligibility(blob)
	switch elig {
	case models.EligGlobal:
		score += 15
		why = append(why, "global-friendly")
	case models.EligUSOnly:
		score -= 30
		why = append(why, "US-only")
	}

	if strings.Contains(blob, "react native") || strings.Contains(blob, "expo") {
		score += 15
		why = append(why, "react native")
	}

	score = clamp(score, 0, 100)
	reason := "weak match"
	if len(why) > 0 {
		reason = strings.Join(why, " · ")
	}
	return score, reason, elig
}

func detectEligibility(blob string) models.Eligibility {
	for _, s := range profile.USOnlySignals {
		if strings.Contains(blob, s) {
			return models.EligUSOnly
		}
	}
	for _, s := range profile.GlobalSignals {
		if strings.Contains(blob, s) {
			return models.EligGlobal
		}
	}
	return models.EligUnknown
}

// Refine asks Claude to re-judge the strongest candidates. No-op without a key.
type verdict struct {
	ID          string `json:"id"`
	Score       int    `json:"score"`
	Eligibility string `json:"eligibility"`
	Reason      string `json:"reason"`
}

func Refine(ctx context.Context, c *llm.Client, jobs []models.Job) map[string]struct {
	Score       int
	Eligibility models.Eligibility
	Reason      string
} {
	result := map[string]struct {
		Score       int
		Eligibility models.Eligibility
		Reason      string
	}{}
	if !c.Enabled() || len(jobs) == 0 {
		return result
	}

	type item struct {
		ID, Title, Company, Location, Snippet string
	}
	list := make([]item, 0, len(jobs))
	for _, j := range jobs {
		list = append(list, item{j.ID, j.Title, j.Company, j.Location, truncate(j.Description, 500)})
	}
	payload, _ := json.Marshal(list)

	prompt := fmt.Sprintf(`You score job posts for a candidate. Return ONLY a JSON array, no prose.

Candidate: %s, based in %s. Final-year CS student.
Wants: remote Frontend / Mobile (React, React Native, Expo, Next.js, TypeScript) roles at funded startups.
Learning Go backend. Junior / intern / SDE-1 level. CANNOT take US-work-authorization-only roles.

For each job return {"id","score" (0-100 fit),"eligibility" ("global"|"us-only"|"unknown"),"reason" (<=12 words)}.
Penalize senior and US-only roles heavily.

Jobs:
%s`, profile.Name, profile.BasedIn, string(payload))

	text, err := c.Complete(ctx, prompt, 1500)
	if err != nil {
		fmt.Println("llm refine:", err)
		return result
	}
	start, end := strings.Index(text, "["), strings.LastIndex(text, "]")
	if start < 0 || end <= start {
		return result
	}
	var verdicts []verdict
	if err := json.Unmarshal([]byte(text[start:end+1]), &verdicts); err != nil {
		fmt.Println("llm parse:", err)
		return result
	}
	for _, v := range verdicts {
		result[v.ID] = struct {
			Score       int
			Eligibility models.Eligibility
			Reason      string
		}{v.Score, models.Eligibility(v.Eligibility), v.Reason}
	}
	return result
}

// ---- small helpers ----

func firstHit(hay string, needles []string) string {
	for _, n := range needles {
		if strings.Contains(hay, n) {
			return n
		}
	}
	return ""
}
func allHits(hay string, needles []string) []string {
	var out []string
	for _, n := range needles {
		if strings.Contains(hay, n) {
			out = append(out, n)
		}
	}
	return out
}
func firstN(s []string, n int) []string {
	if len(s) < n {
		return s
	}
	return s[:n]
}
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
