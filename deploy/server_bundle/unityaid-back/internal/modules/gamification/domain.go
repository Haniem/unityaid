package gamification

import "time"

type Achievement struct {
	ID           string    `json:"id"`
	Code         string    `json:"code"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Icon         string    `json:"icon"`
	PointsReward int       `json:"pointsReward"`
	CreatedAt    time.Time `json:"createdAt"`
}

type UserAchievement struct {
	Achievement
	EarnedAt time.Time `json:"earnedAt"`
}

type Profile struct {
	UserID          string            `json:"userId"`
	UserName        string            `json:"userName"`
	Email           string            `json:"email"`
	AvatarURL       *string           `json:"avatarUrl"`
	TotalHours      float64           `json:"totalHours"`
	Points          int               `json:"points"`
	Level           int               `json:"level"`
	NextLevelAt     int               `json:"nextLevelAt"`
	LevelProgress   float64           `json:"levelProgress"`
	Achievements    []UserAchievement `json:"achievements"`
	NextAchievement *Achievement      `json:"nextAchievement"`
}

type LeaderboardEntry struct {
	Rank       int     `json:"rank"`
	UserID     string  `json:"userId"`
	UserName   string  `json:"userName"`
	Email      string  `json:"email"`
	AvatarURL  *string `json:"avatarUrl"`
	TotalHours float64 `json:"totalHours"`
	Points     int     `json:"points"`
	Level      int     `json:"level"`
}

type MeResponse struct {
	Item Profile `json:"item"`
}

type AchievementsResponse struct {
	Items []Achievement `json:"items"`
}

type LeaderboardResponse struct {
	Items []LeaderboardEntry `json:"items"`
}
