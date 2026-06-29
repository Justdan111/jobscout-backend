package apply

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"jobscout/internal/llm"
	"jobscout/internal/models"
	"jobscout/internal/profile"
)

// Draft is the generated application material for one job.
type Draft struct {
	ResumeHighlights string `json:"resumeHighlights"`
	CoverEmail       string `json:"coverEmail"`
	Pitch            string `json:"pitch"`
}

// Generate writes tailored application material for a job, grounded in Dan's
// real resume summary. Requires an Anthropic key.
func Generate(ctx context.Context, c *llm.Client, job models.Job) (Draft, error) {
	if !c.Enabled() {
		return Draft{}, errors.New("ANTHROPIC_API_KEY not set — application drafting is disabled")
	}

	prompt := fmt.Sprintf(`You are helping a candidate apply for a job. Write honest, specific, non-generic material. No clichés ("I am thrilled"), no invented achievements — use ONLY facts from the resume. Reorder and emphasize; never fabricate. Keep it tight and human.

=== CANDIDATE RESUME ===
%s

=== JOB ===
Title: %s
Company: %s
Location: %s
Details: %s

=== TASK ===
Return ONLY a JSON object, no prose, no markdown:
{
  "resumeHighlights": "A tailored resume header for THIS job: a 2-sentence professional summary aimed at this role, then 4-5 bullet lines (each prefixed with '• ') that reorder and emphasize the candidate's most relevant real projects/skills for this job. Facts only — pulled from the resume above.",
  "coverEmail": "A complete cover email (first line 'Subject: ...', then the body). 130-180 words. Address the specific company/role, connect 2-3 of the candidate's real projects/skills to what this job needs, close with a clear call to talk. Sign off as Emmanuel Ngulube with portfolio dantech-xi.vercel.app and GitHub github.com/Justdan111.",
  "pitch": "A 2-3 sentence message for a quick application box or DM — punchy, specific to this role."
}`, profile.ResumeSummary, job.Title, job.Company, job.Location, truncate(job.Description, 900))

	text, err := c.Complete(ctx, prompt, 1600)
	if err != nil {
		return Draft{}, err
	}
	start, end := strings.Index(text, "{"), strings.LastIndex(text, "}")
	if start < 0 || end <= start {
		return Draft{}, fmt.Errorf("unexpected model output")
	}
	var d Draft
	if err := json.Unmarshal([]byte(text[start:end+1]), &d); err != nil {
		return Draft{}, err
	}
	return d, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
