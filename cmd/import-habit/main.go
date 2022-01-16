package main

import (
	"fmt"
	"github.com/JoelOtter/dayligo"
	"github.com/JoelOtter/dayligo/cmd/import-habit/internal"
	"github.com/JoelOtter/dayligo/cmd/import-habit/internal/habit"
	"github.com/spf13/cobra"
)

func main() {
	var daylioFilePath string
	var habitFilePath string
	var outputFilePath string

	cmd := &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			backup, err := dayligo.ReadBackupFromFile(daylioFilePath)
			if err != nil {
				return err
			}
			defer func() {
				if err := backup.Close(); err != nil {
					fmt.Printf("Failed to close backup file: %s\n", err.Error())
				}
			}()
			fmt.Printf("Loaded %d day entries, %d goals, %d goal entries.\n", len(backup.DayEntries), len(backup.Goals), len(backup.GoalEntries))
			habits, err := habit.ReadEntriesFromFile(habitFilePath)
			if err != nil {
				return err
			}
			for habitName, entries := range habits {
				fmt.Printf("Found habit \"%s\" with %d entries.\n", habitName, len(entries))
			}

			if err := internal.ImportGoalsFromHabit(backup, habits); err != nil {
				return fmt.Errorf("failed to import goals from habit: %w", err)
			}

			if outputFilePath == "" {
				fmt.Println("No output path provided; all done.")
				return nil
			}

			if err := backup.WriteToFile(outputFilePath); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.PersistentFlags().StringVar(
		&daylioFilePath,
		"daylio-file-path",
		"",
		"The path to the .daylio file",
	)
	_ = cmd.MarkPersistentFlagRequired("daylio-file-path")

	cmd.PersistentFlags().StringVar(
		&habitFilePath,
		"habit-file-path",
		"",
		"The path to the Habit-exported CSV file",
	)
	_ = cmd.MarkPersistentFlagRequired("habit-file-path")

	cmd.PersistentFlags().StringVar(
		&outputFilePath,
		"output-file-path",
		"",
		"The path to the desired output file. If empty, dry-run only.",
	)

	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
