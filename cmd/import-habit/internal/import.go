package internal

import (
	"fmt"
	"github.com/JoelOtter/dayligo"
	"sort"
	"time"
)

func ImportGoalsFromHabit(backup *dayligo.Backup, habitEntries map[string][]time.Time) error {
	goalIDMapping, err := getHabitToGoalIDMapping(backup, habitEntries)
	if err != nil {
		return fmt.Errorf("failed to get goal-habit mapping: %w", err)
	}
	// ID ordering doesn't appear to matter. Goal entries are ordered by date,
	// newest-first, with all goals mixed together.
	var goalEntries []dayligo.GoalEntry
	i := int64(1)
	for habit, goal := range goalIDMapping {
		for _, date := range habitEntries[habit] {
			// Set goal entries to 11pm to align ish with when I do the diary.
			t := date.Add(23 * time.Hour)
			goalEntries = append(goalEntries, dayligo.GoalEntry{
				CreatedAt: t.UnixMilli(),
				Hour:      int64(t.Hour()),
				Minute:    int64(t.Minute()),
				Second:    int64(t.Second()),
				Day:       int64(t.Day()),
				Month:     int64(t.Month()),
				Year:      int64(t.Year()),
				GoalID:    goal,
				ID:        i,
			})
			i += 1
		}
	}
	sort.Slice(goalEntries, func(i, j int) bool {
		return goalEntries[i].CreatedAt < goalEntries[j].CreatedAt
	})

	// Set goal start dates to earliest Habit entry
	for habit, dates := range habitEntries {
		ds := dates
		sort.Slice(ds, func(i, j int) bool {
			return ds[i].Before(ds[j])
		})
		for i, goal := range backup.Goals {
			if goal.ID == goalIDMapping[habit] {
				goal.CreatedAt = ds[0].UnixMilli()
				backup.Goals[i] = goal
				break
			}
		}
	}

	backup.GoalEntries = goalEntries

	return nil
}

func getHabitToGoalIDMapping(backup *dayligo.Backup, habitEntries map[string][]time.Time) (map[string]int64, error) {
	result := make(map[string]int64)
	for habitEntry := range habitEntries {
		id := getGoalIDForHabit(habitEntry, backup)
		if id == 0 {
			return nil, fmt.Errorf("failed to get goal ID for habit %s", habitEntry)
		}
		result[habitEntry] = id
	}
	return result, nil
}

func getGoalIDForHabit(habitName string, backup *dayligo.Backup) int64 {
	for _, goal := range backup.Goals {
		if goal.Name == habitName {
			return goal.ID
		}
	}
	// No matching goal name - try matching on tag
	tagID := int64(0)
	for _, tag := range backup.Tags {
		if habitName == tag.Name {
			tagID = tag.ID
			break
		}
	}
	if tagID == 0 {
		return 0
	}
	for _, goal := range backup.Goals {
		if goal.TagID == tagID {
			return goal.ID
		}
	}
	return 0
}
