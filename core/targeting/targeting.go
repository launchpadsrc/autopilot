package targeting

import (
	"log/slog"
	"slices"
	"strings"

	"github.com/samber/lo"

	"launchpad.icu/autopilot/core/cvschema"
	"launchpad.icu/autopilot/core/launchpad"
	"launchpad.icu/autopilot/internal/database"
)

var logger = slog.With("package", "core/targeting")

// MinScore is a recommended minimum score for targeting.
const MinScore = 5

type FindParams struct {
	Profile launchpad.UserProfile // required
	Resume  cvschema.Resume       // required
	Jobs    []database.Job        // required
	// MinScore allows filtering out jobs that have a score below the specified value.
	MinScore int
}

// Job represents a job that matches the user's profile and resume.
type Job struct {
	database.Job
	// Matches contains the keywords that matched the job.
	Matches []string `json:"matches"`
	// Score is the sum of weights of the matched keywords.
	Score int `json:"score"`
}

// Find returns a list of jobs that match the user's profile and resume.
// It filters jobs based on the user's seniority and skills, and scores them based on the matched keywords.
// Skill keywords from the profile are given more weight than those from the resume.
// The jobs are sorted by score in descending order.
func Find(params FindParams) (targeted []Job, _ error) {
	if params.MinScore == 0 {
		params.MinScore = 1
	}
	return find(params)
}

func find(params FindParams) (targeted []Job, _ error) {
	var (
		seniority = params.ProfileSeniority()
		keywords  = params.Keywords()
		keys      = lo.Keys(keywords)
	)

	logger.Debug("found keywords", "keywords", keywords)

	for _, job := range params.Jobs {
		// Skip jobs that are not in the user's seniority range.
		if !slices.Contains(seniority, normalize(job.SeniorityAI)) {
			continue
		}

		hashtags := lo.Map(job.HashtagsAI, func(h string, _ int) string {
			return keywordReplacer.Replace(h)
		})

		matches := lo.Intersect(hashtags, keys)
		// Skip jobs that do not match the user's roles.
		if len(matches) == 0 {
			continue
		}

		score := lo.Sum(lo.Map(matches, func(k string, _ int) int {
			return keywords[k]
		}))

		if score < params.MinScore {
			continue
		}

		targeted = append(targeted, Job{
			Job:     job,
			Matches: matches,
			Score:   score,
		})
	}

	slices.SortFunc(targeted, func(a, b Job) int {
		return b.Score - a.Score
	})

	return targeted, nil
}

// ProfileSeniority returns a slice of seniority levels that are less than or equal to the user's seniority.
func (params FindParams) ProfileSeniority() []string {
	i := slices.Index(seniorityRange, normalize(params.Profile.Seniority))
	if i == -1 {
		logger.Warn("unknown seniority level", "seniority", params.Profile.Seniority)
		return seniorityRange // return all levels if unknown
	}
	return seniorityRange[:i]
}

// Keywords returns a map of keywords with their weights based on the user's profile and resume.
func (params FindParams) Keywords() map[string]int {
	var profileKeywords []keyword
	for _, skill := range params.Profile.Stack {
		profileKeywords = append(profileKeywords, keyword{
			kw:     normalizeKeyword(skill.Tech),
			weight: skill.Level * 2,
		})
	}

	var resumeKeywords []keyword
	for _, skill := range params.Resume.Skills {
		for _, kw := range skill.Keywords {
			resumeKeywords = append(resumeKeywords, keyword{
				kw:     normalizeKeyword(kw),
				weight: 1,
			})
		}
	}

	return joinedKeywords(profileKeywords, resumeKeywords)
}

var seniorityRange = []string{
	"trainee",
	"junior",
	"middle",
	"senior",
}

var keywordReplacer = strings.NewReplacer(
	"go", "golang",
	"js", "javascript",
)

type keyword struct {
	kw     string
	weight int // 1..5
}

// joinedKeywords merges multiple slices of keywords into a single map.
func joinedKeywords(kwss ...[]keyword) map[string]int {
	joined := make(map[string]int)
	for _, kws := range kwss {
		for _, kw := range kws {
			weight, exists := joined[kw.kw]
			if !exists || kw.weight > weight {
				joined[kw.kw] = kw.weight
			}
		}
	}
	return joined
}

func normalize(s string) string {
	return strings.TrimSpace(strings.ToLower(s))
}

func normalizeKeyword(kw string) string {
	// 1. Remove punctuation.
	kw = strings.Trim(kw, ".,;:!?\"'()[]{}<>/\\|~@#$%^&*_-+=")
	// 2. Only first word.
	kw = strings.Split(kw, " ")[0]
	// 3. Replace common abbreviations.
	kw = keywordReplacer.Replace(kw)
	// 4. Basic normalization.
	return normalize(kw)
}
