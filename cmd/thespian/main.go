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

var actorCmd = &cobra.Command{
	Use:        "actor [flags] TypeName",
	Short:      "Generate an actor definition in this package with the given name",
	Args:       cobra.ExactArgs(1),
	ArgAliases: []string{"typeName"},
	Run: func(cmd *cobra.Command, args []string) {
		typeName := args[0]
		gen.GenerateActor(typeName)
	},
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("thespian: ")

	rootCmd.AddCommand(actorCmd)

	cobra.CheckErr(rootCmd.Execute())
}
