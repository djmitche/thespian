package main

import (
	"log"

	"github.com/djmitche/thespian/internal/gen"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "thespian",
	Short: "Thespian code-generation tool",
}

var generateCmd = &cobra.Command{
	Use:   "generate [flags]",
	Short: "Generate type definitions based on thespian.yml in this directory",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		gen.Generate()
	},
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("thespian: ")

	rootCmd.AddCommand(generateCmd)

	cobra.CheckErr(rootCmd.Execute())
}
