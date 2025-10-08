package calculations

import (
	"main/internal/leaderboard/model/leaderboard"
	"sort"
)

func Sum(arr []int) int {
	total := 0
	for _, v := range arr {
		total += v
	}
	return total
}

func SortLeaderboards(leaderboards []leaderboard.Leaderboard) []leaderboard.Leaderboard {
	sort.Slice(leaderboards, func(i, j int) bool {
		if leaderboards[i].WinCount == leaderboards[j].WinCount {
			scoreI := Sum(leaderboards[i].Scores)
			scoreJ := Sum(leaderboards[j].Scores)
			return scoreI > scoreJ
		}
		return leaderboards[i].WinCount > leaderboards[j].WinCount
	})

	return leaderboards
}
