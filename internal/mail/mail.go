package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"sort"
	"time"

	"jobscout/internal/config"
	"jobscout/internal/models"
)

// SendDigest emails the strongest new, eligible jobs. No-ops (with a log) if
// email isn't configured.
func SendDigest(ctx context.Context, cfg config.Config, jobs []models.Job) (sent int, err error) {
	top := make([]models.Job, 0)
	for _, j := range jobs {
		if j.Status == models.StatusNew && j.Eligibility != models.EligUSOnly {
			top = append(top, j)
		}
	}
	sort.Slice(top, func(i, k int) bool { return top[i].Score > top[k].Score })
	if len(top) > 12 {
		top = top[:12]
	}

	if cfg.ResendKey == "" || cfg.DigestTo == "" {
		fmt.Println("email not configured (RESEND_API_KEY / DIGEST_TO); skipping")
		return 0, nil
	}
	if len(top) == 0 {
		return 0, nil
	}

	body, _ := json.Marshal(map[string]any{
		"from":    cfg.DigestFrom,
		"to":      cfg.DigestTo,
		"subject": fmt.Sprintf("JobScout: %d new matches", len(top)),
		"html":    render(top),
	})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.resend.com/emails", bytes.NewReader(body))
	req.Header.Set("content-type", "application/json")
	req.Header.Set("authorization", "Bearer "+cfg.ResendKey)

	client := &http.Client{Timeout: 20 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return 0, fmt.Errorf("resend status %d", res.StatusCode)
	}
	return len(top), nil
}

func render(jobs []models.Job) string {
	var rows bytes.Buffer
	for _, j := range jobs {
		rows.WriteString(fmt.Sprintf(`<tr><td style="padding:14px 0;border-bottom:1px solid #E6E8EC;">
<div style="font:600 15px sans-serif;color:#15171C;"><a href="%s" style="color:#117C6F;text-decoration:none;">%s</a></div>
<div style="font:400 13px sans-serif;color:#6B7280;margin-top:2px;">%s · %s · score %d · %s</div>
<div style="font:400 12px sans-serif;color:#9097A1;margin-top:3px;">%s</div></td></tr>`,
			html.EscapeString(j.URL), html.EscapeString(j.Title),
			html.EscapeString(j.Company), html.EscapeString(j.Location),
			j.Score, j.Eligibility, html.EscapeString(j.Reason)))
	}
	return fmt.Sprintf(`<div style="max-width:560px;margin:0 auto;font-family:sans-serif;">
<h1 style="font:700 18px sans-serif;color:#15171C;">Today's matches</h1>
<p style="color:#6B7280;font-size:13px;">Sorted by fit. US-only roles filtered out.</p>
<table style="width:100%%;border-collapse:collapse;">%s</table></div>`, rows.String())
}
