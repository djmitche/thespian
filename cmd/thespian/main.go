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
	Short:      "Generate an actor definition in this package based on the private type with the given name",
	Args:       cobra.ExactArgs(1),
	ArgAliases: []string{"typeName"},
	Run: func(cmd *cobra.Command, args []string) {
		typeName := args[0]
		gen.GenerateActor(typeName)
	},
}

var mailboxCmd = &cobra.Command{
	Use:        "mailbox [flags] TypeName",
	Short:      "Generate an mailbox definition in this package based on the private type with the given name",
	Args:       cobra.ExactArgs(1),
	ArgAliases: []string{"typeName"},
	Run: func(cmd *cobra.Command, args []string) {
		typeName := args[0]
		gen.GenerateMailbox(typeName)
	},
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("thespian: ")

	rootCmd.AddCommand(actorCmd)
	rootCmd.AddCommand(mailboxCmd)

	cobra.CheckErr(rootCmd.Execute())
}
