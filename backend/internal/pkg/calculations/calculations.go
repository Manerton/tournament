package calculations

import (
	"main/internal/leaderboard/model/leaderboard"
	"main/internal/team/model/team"
	"math"
	"sort"
)

func Sum(arr []int32) int32 {
	var total = int32(0)
	for _, v := range arr {
		total += v
	}
	return total
}

func SortLeaderboards(leaderboards []leaderboard.Leaderboard) map[int][]leaderboard.Leaderboard {
	winGroup := map[int][]leaderboard.Leaderboard{}
	// Группируем команды по количеству побед
	for _, lb := range leaderboards {
		winGroup[lb.WinCount] = append(winGroup[lb.WinCount], lb)
	}

	// Сортируем команды внутри каждой группы по сумме очков
	for winCount, group := range winGroup {
		sort.Slice(group, func(i, j int) bool {
			scoreI := Sum(group[i].Scores)
			scoreJ := Sum(group[j].Scores)
			return scoreI > scoreJ
		})
		winGroup[winCount] = group
	}

	// sort.Slice(leaderboards, func(i, j int) bool {
	// 	if leaderboards[i].WinCount == leaderboards[j].WinCount {
	// 		scoreI := Sum(leaderboards[i].Scores)
	// 		scoreJ := Sum(leaderboards[j].Scores)
	// 		return scoreI > scoreJ
	// 	}
	// 	return leaderboards[i].WinCount > leaderboards[j].WinCount
	// })

	// return leaderboards

	return winGroup
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
