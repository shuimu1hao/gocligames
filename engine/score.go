package engine

import (
	"encoding/json"
	"os"
	"sort"
)

// ScoreRec 排行榜一条记录。
type ScoreRec struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
	Extra string `json:"extra,omitempty"`
}

// Scoreboard JSON 排行榜（TopN 持久化，进程退出后保留）。
type Scoreboard struct {
	Path   string
	Limit  int
	Scores []ScoreRec
}

// NewScoreboard 加载（或创建）排行榜。
func NewScoreboard(path string, limit int) *Scoreboard {
	sb := &Scoreboard{Path: path, Limit: limit}
	sb.Scores = sb.load()
	return sb
}

func (sb *Scoreboard) load() []ScoreRec {
	data, err := os.ReadFile(sb.Path)
	if err != nil {
		return nil
	}
	var recs []ScoreRec
	if json.Unmarshal(data, &recs) != nil {
		return nil
	}
	return recs
}

func (sb *Scoreboard) save() {
	data, err := json.MarshalIndent(sb.Scores, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(sb.Path, data, 0o644)
}

// Add 添加一条记录，自动排序并截断到 TopN，返回名次（0 起）；未进榜返回 -1。
func (sb *Scoreboard) Add(name string, score int, extra string) int {
	rec := ScoreRec{Name: name, Score: score, Extra: extra}
	sb.Scores = append(sb.Scores, rec)
	sort.SliceStable(sb.Scores, func(i, j int) bool {
		return sb.Scores[i].Score > sb.Scores[j].Score
	})
	if len(sb.Scores) > sb.Limit {
		sb.Scores = sb.Scores[:sb.Limit]
	}
	sb.save()
	for i := range sb.Scores {
		if sb.Scores[i].Score == rec.Score && sb.Scores[i].Name == rec.Name && sb.Scores[i].Extra == rec.Extra {
			return i
		}
	}
	return -1
}

// Top 返回前 n 条记录（n <= 0 时返回全部）。
func (sb *Scoreboard) Top(n int) []ScoreRec {
	if n <= 0 || n > len(sb.Scores) {
		n = len(sb.Scores)
	}
	return sb.Scores[:n]
}
