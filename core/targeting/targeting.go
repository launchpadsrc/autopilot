package targeting

import (
	"log/slog"
	"slices"
	"strings"

	"github.com/samber/lo"

	"launchpad.icu/autopilot/core/cvschema"
	"launchpad.icu/autopilot/core/launchpad"
	"launchpad.icu/autopilot/database/sqlc"
)

var logger = slog.With("package", "core/targeting")

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

// TargetedJob represents a job that matches the user's profile and resume.
type TargetedJob struct {
	sqlc.Job
	// Matches contains the keywords that matched the job.
	Matches []string `json:"matches"`
	// Score is the sum of weights of the matched keywords.
	Score int `json:"score"`
}

// Find returns a list of jobs that match the user's profile and resume.
// It filters jobs based on the user's seniority and skills, and scores them based on the matched keywords.
// Skill keywords from the profile are given more weight than those from the resume.
// The jobs are sorted by score in descending order.
func Find(profile launchpad.UserProfile, resume cvschema.Resume, jobs []sqlc.Job) (targeted []TargetedJob, _ error) {
	profileSkills := profileSkillKeywords(profile)
	resumeSkills := resumeSkillKeywords(resume)

	joinedSkills := joinedKeywords(profileSkills, resumeSkills)
	logger.Debug("found skills", "keywords", joinedSkills)

	seniorities := profileSeniorities(profile)
	keywords := lo.Keys(joinedSkills)

	for _, job := range jobs {
		// Skip jobs that are not in the user's seniority range.
		if !slices.Contains(seniorities, normalize(job.SeniorityAI)) {
			continue
		}

		hashtags := lo.Map(job.HashtagsAI, func(h string, _ int) string {
			return keywordReplacer.Replace(h)
		})

		matches := lo.Intersect(hashtags, keywords)
		// Skip jobs that do not match the user's roles.
		if len(matches) == 0 {
			continue
		}

		score := lo.Sum(lo.Map(matches, func(k string, _ int) int {
			return joinedSkills[k]
		}))

		targeted = append(targeted, TargetedJob{
			Job:     job,
			Matches: matches,
			Score:   score,
		})
	}

	slices.SortFunc(targeted, func(a, b TargetedJob) int {
		return b.Score - a.Score
	})

	return targeted, nil
}

type keyword struct {
	kw     string
	weight int // 1..5
}

func profileSeniorities(profile launchpad.UserProfile) []string {
	i := slices.Index(seniorityRange, normalize(profile.Seniority))
	return slices.Clone(seniorityRange[:i])
}

// profileSkillKeywords extracts keywords from the user profile skills section.
// All the keywords will have a weight of 2*level for additional priority.
func profileSkillKeywords(profile launchpad.UserProfile) (kws []keyword) {
	for _, skill := range profile.Stack {
		kws = append(kws, keyword{
			kw:     normalizeKeyword(skill.Tech),
			weight: skill.Level * 2,
		})
	}
	return kws
}

// resumeSkillKeywords extracts keywords from the resume skills section.
// All the keywords will have a weight of 1.
func resumeSkillKeywords(resume cvschema.Resume) (kws []keyword) {
	for _, skill := range resume.Skills {
		for _, kw := range skill.Keywords {
			kws = append(kws, keyword{
				kw:     normalizeKeyword(kw),
				weight: 1,
			})
		}
	}
	return kws
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
