package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/tylerkeyes/jot-me/internal/database"
	"github.com/tylerkeyes/jot-me/internal/render"
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

	viewCmd = &cobra.Command{
		Use:   "view",
		Short: "View your notes.",
		Args:  cobra.NoArgs,
		Run:   viewNotes,
	}
)

func writeNote(cmd *cobra.Command, args []string) error {
	err := db.WriteNote(noteGroup, strings.Join(args, " "))
	if err != nil {
		fmt.Printf("problem writing the note: %v\n", err)
		return err
	}

	return nil
}

func viewNotes(cmd *cobra.Command, args []string) {
	if noteGroup == "" {
		noteGroup = "general"
	}
	model := render.InitialModel(&db, noteGroup)
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Printf("that ain't gonna fly: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// initialize cobra command handlers
	if Version == "" {
		Version = "unknown (built from source)"
	}
	rootCmd.Version = Version
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	noteCmd.Flags().StringVarP(&noteGroup, "group", "g", "", "choose which group to write the note.")
	viewCmd.Flags().StringVarP(&noteGroup, "group", "g", "", "choose which group to read from.")

	rootCmd.AddCommand(noteCmd, viewCmd)

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
