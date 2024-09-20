package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tylerkeyes/jot-me/internal/database"
)

var (
	Version = "0.0.1"

	db        database.Service
	noteGroup string

	rootCmd = &cobra.Command{
		Use:   "jot",
		Short: "jot-me, a quick note taking app.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	noteCmd = &cobra.Command{
		Use:   "note",
		Short: "Write a note.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  writeNote,
	}
)

func writeNote(cmd *cobra.Command, args []string) error {
	err := db.WriteNote(noteGroup, strings.Join(args, " "))
	if err != nil {
		fmt.Printf("problem writing the note: %v\n", err)
	}

	return nil
}

func init() {
	// initialize cobra command handlers
	if Version == "" {
		Version = "unknown (built from source)"
	}
	rootCmd.Version = Version
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	noteCmd.Flags().StringVarP(&noteGroup, "group", "g", "", "choose which group to write the note")

	rootCmd.AddCommand(noteCmd)

	// initialize the local db
	db = database.New()
}

func main() {
	defer db.Close()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}
