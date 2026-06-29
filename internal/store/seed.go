package store

import (
	"time"

	"jobscout/internal/models"
)

func seedJobs() []models.Job {
	now := time.Now().UTC().Format(time.RFC3339)
	return []models.Job{
		{
			ID: "seed-whip-rn", Source: "seed",
			Title: "Founding Mobile Engineer (React Native)", Company: "Whip (YC W24)",
			Location: "Remote", URL: "https://www.ycombinator.com/companies/whip/jobs",
			Description: "Build a social feed for interactive AI creations in React Native. Early team, own mobile end to end.",
			Tags:        []string{"react native", "mobile", "typescript"},
			PostedAt:    now, Score: 88, Eligibility: models.EligUnknown,
			Reason: "Exact stack (React Native). 'Founding' leans senior — pitch portfolio hard.",
			YC:     true, YCBatch: "Winter 2024", Funded: true, NewlyFunded: true,
			Status: models.StatusNew, DiscoveredAt: now,
		},
		{
			ID: "seed-coulomb-fe", Source: "seed",
			Title: "Frontend Developer — SDE 1", Company: "Coulomb AI (YC S21)",
			Location: "Remote", URL: "https://www.ycombinator.com/companies/coulomb-ai/jobs",
			Description: "Battery observability for EVs. SDE-1 frontend role building dashboards and data tooling.",
			Tags:        []string{"frontend", "react", "sde 1"},
			PostedAt:    now, Score: 82, Eligibility: models.EligUnknown,
			Reason: "SDE-1 = junior friendly. React frontend matches your core strength.",
			YC:     true, YCBatch: "Summer 2021", Funded: true, NewlyFunded: true,
			Status: models.StatusNew, DiscoveredAt: now,
		},
		{
			ID: "seed-traceroot-intern", Source: "seed",
			Title: "Software Engineering Intern (Full Stack)", Company: "TraceRoot.AI (YC S25)",
			Location: "San Francisco / Remote", URL: "https://www.ycombinator.com/companies/traceroot-ai/jobs",
			Description: "Open-source self-healing layer for AI agents. Internship, full stack, remote possible.",
			Tags:        []string{"full stack", "intern", "typescript"},
			PostedAt:    now, Score: 74, Eligibility: models.EligUnknown,
			Reason: "Intern framing fits final-year status. Confirm remote outside the US.",
			YC:     true, YCBatch: "Summer 2025", Funded: true, NewlyFunded: true, Internship: true,
			Status: models.StatusNew, DiscoveredAt: now,
		},
	}
}
