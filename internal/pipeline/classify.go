package pipeline

import (
	"strings"

	"jobscout/internal/models"
	"jobscout/internal/yc"
)

var vcSignals = []string{
	"pre-seed", "seed round", "series a", "series b", "series c",
	"backed by", "raised", "venture", "vc-backed", "y combinator",
}

// classify sets the label fields used by the table filters. ycIndex maps
// normalized company name -> batch for recent (newly funded) YC companies.
func classify(j *models.Job, ycIndex map[string]string) {
	blob := strings.ToLower(strings.Join([]string{
		j.Title, j.Company, strings.Join(j.Tags, " "), j.Description,
	}, " "))

	// Internship: explicit source tag or any "intern" mention in the title.
	j.Internship = hasTag(j.Tags, "internship") ||
		strings.Contains(strings.ToLower(j.Title), "intern")

	// YC: from the YC source, or company matches a recent YC company.
	batch, inIndex := ycIndex[yc.Normalize(j.Company)]
	if j.Source == "yc" || inIndex {
		j.YC = true
		j.YCBatch = batch
	}

	// Newly funded: we only track recent YC batches, so any YC match qualifies.
	j.NewlyFunded = j.YC

	// Funded: YC, or a VC/funding signal in the text.
	j.Funded = j.YC
	if !j.Funded {
		for _, s := range vcSignals {
			if strings.Contains(blob, s) {
				j.Funded = true
				break
			}
		}
	}
}

func hasTag(tags []string, want string) bool {
	for _, t := range tags {
		if t == want {
			return true
		}
	}
	return false
}
