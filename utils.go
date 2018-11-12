package main

import (
	"io"
	"log"
	"time"

	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
)

func getStringFlag(cmd *cobra.Command, name string) string {
	arg, err := cmd.Flags().GetString(name)
	if err != nil {
		log.Fatal(err)
	}
	return arg
}
func getStringArrayFlag(cmd *cobra.Command, name string) []string {
	arg, err := cmd.Flags().GetStringArray(name)
	if err != nil {
		log.Fatal(err)
	}
	return arg
}
func getBoolFlag(cmd *cobra.Command, name string) bool {
	arg, err := cmd.Flags().GetBool(name)
	if err != nil {
		log.Fatal(err)
	}
	return arg
}
func getIntFlag(cmd *cobra.Command, name string) int {
	arg, err := cmd.Flags().GetInt(name)
	if err != nil {
		log.Fatal(err)
	}
	return arg
}

func newSpinner(out io.Writer, suffix string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Writer = out
	s.Suffix = " " + suffix
	s.FinalMSG = ""
	s.Color("cyan")
	s.Start()
	return s
}
