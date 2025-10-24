package calculations

import (
	"main/internal/leaderboard/model/leaderboard"
	"main/internal/team/model/team"
	"math"
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

func SortTeamByInitialScores(teams []team.Team) []team.Team {
	sort.Slice(teams, func(i, j int) bool {
		return teams[i].InitialScore > teams[j].InitialScore
	})
	return teams
}

func MaxTours(countTeams int) int {
	result := math.Log2(float64(countTeams))
	return int(math.Floor(result))
}
