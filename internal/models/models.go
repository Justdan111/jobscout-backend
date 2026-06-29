package models

// Eligibility expresses whether Dan can realistically apply from Nigeria.
type Eligibility string

const (
	EligGlobal  Eligibility = "global"
	EligUSOnly  Eligibility = "us-only"
	EligUnknown Eligibility = "unknown"
)

// TriageStatus is the quick decision on a discovered job.
type TriageStatus string

const (
	StatusNew       TriageStatus = "new"
	StatusSaved     TriageStatus = "saved"
	StatusDismissed TriageStatus = "dismissed"
)

// AppStage is the application pipeline stage.
type AppStage string

const (
	StageDrafted   AppStage = "drafted"
	StageApplied   AppStage = "applied"
	StageInterview AppStage = "interviewing"
	StageOffer     AppStage = "offer"
	StageRejected  AppStage = "rejected"
)

// Job is a discovered, scored posting.
type Job struct {
	ID          string      `json:"id"`
	Source      string      `json:"source"`
	Title       string      `json:"title"`
	Company     string      `json:"company"`
	Location    string      `json:"location"`
	URL         string      `json:"url"`
	Description string      `json:"description"`
	Tags        []string    `json:"tags"`
	PostedAt    string      `json:"postedAt"`
	Score       int         `json:"score"`
	Reason      string      `json:"reason"`
	Eligibility Eligibility `json:"eligibility"`
	// classification labels (drive the table filters)
	YC           bool         `json:"yc"`
	YCBatch      string       `json:"ycBatch"`
	Funded       bool         `json:"funded"`
	NewlyFunded  bool         `json:"newlyFunded"`
	Internship   bool         `json:"internship"`
	Status       TriageStatus `json:"status"`
	DiscoveredAt string       `json:"discoveredAt"`
}

// Application tracks Dan's progress on a specific job and holds the draft.
type Application struct {
	JobID            string   `json:"jobId"`
	Stage            AppStage `json:"stage"`
	ResumeHighlights string   `json:"resumeHighlights"`
	CoverEmail       string   `json:"coverEmail"`
	Pitch            string   `json:"pitch"`
	Notes            string   `json:"notes"`
	CreatedAt        string   `json:"createdAt"`
	UpdatedAt        string   `json:"updatedAt"`
}
