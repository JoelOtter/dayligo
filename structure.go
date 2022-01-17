package dayligo

import (
	"encoding/json"
)

type Backup struct {
	DayEntries       []DayEntry        `json:"dayEntries"`
	Goals            []Goal            `json:"goals"`
	GoalEntries      []GoalEntry       `json:"goalEntries"`
	GoalSuccessWeeks []GoalSuccessWeek `json:"goalSuccessWeeks"`
	Tags             []Tag             `json:"tags"`
	Version          int64             `json:"version"`

	rawMap      map[string]json.RawMessage
	tempDirPath string
}

type DayEntry struct {
	DateTime       int64   `json:"datetime"`
	Mood           int64   `json:"mood"`
	Note           string  `json:"note"`
	Tags           []int64 `json:"tags"`
	Hour           int64   `json:"hour"`
	Minute         int64   `json:"minute"`
	Day            int64   `json:"day"`
	Month          int64   `json:"month"`
	Year           int64   `json:"year"`
	TimeZoneOffset int64   `json:"timeZoneOffset"`

	// Not parsed
	Assets []json.RawMessage `json:"assets"`
}

type GoalEntry struct {
	CreatedAt int64 `json:"createdAt"`
	Hour      int64 `json:"hour"`
	Minute    int64 `json:"minute"`
	Second    int64 `json:"second"`
	Day       int64 `json:"day"`
	Month     int64 `json:"month"`
	Year      int64 `json:"year"`
	GoalID    int64 `json:"goalId"`
	ID        int64 `json:"id"`
}

type Goal struct {
	CreatedAt       int64  `json:"created_at"`
	ID              int64  `json:"goal_id"`
	AvatarID        int64  `json:"id_avatar"`
	ChallengeID     int64  `json:"id_challenge"`
	IconID          int64  `json:"id_icon"`
	TagID           int64  `json:"id_tag"`
	Name            string `json:"name"`
	OrderNumber     int64  `json:"order_number"`
	ReminderEnabled bool   `json:"reminder_enabled"`
	ReminderHour    int64  `json:"reminder_hour"`
	ReminderMinute  int64  `json:"reminder_minute"`
	RepeatType      int64  `json:"repeat_type"`
	RepeatValue     int64  `json:"repeat_value"`
	State           int64  `json:"state"`
}

type GoalSuccessWeek struct {
	CreateAtDay   int64 `json:"create_at_day"`
	CreateAtMonth int64 `json:"create_at_month"`
	CreateAtYear  int64 `json:"create_at_year"`
	GoalID        int64 `json:"goal_id"`
	Week          int64 `json:"week"`
	Year          int64 `json:"year"`
}

type Tag struct {
	CreatedAt int64  `json:"createdAt"`
	Icon      int64  `json:"icon"`
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Order     int64  `json:"order"`
	State     int64  `json:"state"`
}
