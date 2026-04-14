package tenders

import (
	"strconv"
	"strings"
	"time"
)

type ScoreInput struct {
	CompanyRegion   string
	CompanyKeywords []string
}

type ScoreResult struct {
	Score   int      `json:"score"`
	Reasons []string `json:"reasons"`
}

func ScoreTender(t Tender, input ScoreInput) ScoreResult {
	score := 40
	reasons := []string{"base_match=40"}

	if strings.TrimSpace(input.CompanyRegion) != "" && strings.EqualFold(strings.TrimSpace(input.CompanyRegion), strings.TrimSpace(t.Region)) {
		score += 20
		reasons = append(reasons, "region_match=+20")
	}

	matchedKeywords := 0
	searchText := strings.ToLower(t.Title + " " + t.Description)
	for _, keyword := range input.CompanyKeywords {
		kw := strings.ToLower(strings.TrimSpace(keyword))
		if kw == "" {
			continue
		}
		if strings.Contains(searchText, kw) {
			matchedKeywords++
		}
	}
	if matchedKeywords > 0 {
		boost := matchedKeywords * 8
		if boost > 24 {
			boost = 24
		}
		score += boost
		reasons = append(reasons, "keyword_match=+"+strconv.Itoa(boost))
	}

	if t.ClosingAt != nil {
		hoursToClose := t.ClosingAt.Sub(time.Now().UTC()).Hours()
		switch {
		case hoursToClose >= 24 && hoursToClose <= 168:
			score += 10
			reasons = append(reasons, "closing_window_optimal=+10")
		case hoursToClose < 24:
			score -= 8
			reasons = append(reasons, "closing_too_soon=-8")
		}
	}

	if strings.TrimSpace(t.Description) != "" {
		score += 6
		reasons = append(reasons, "rich_description=+6")
	}

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return ScoreResult{
		Score:   score,
		Reasons: reasons,
	}
}
