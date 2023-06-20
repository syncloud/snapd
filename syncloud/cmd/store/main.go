package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/syncloud/store/pkg"
	"os"
)

func main() {
	var rootCmd = &cobra.Command{
		Use: "store",
	}

	var cmdStart = &cobra.Command{
		Use:   "start",
		Short: "Start Syncloud Store",
		Args:  cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			store := pkg.NewSyncloudStore()
			api := pkg.NewApi(store)
			go func() { _ = api.Start() }()
			return store.Start()
		},
	}

	rootCmd.AddCommand(cmdStart)
	err := rootCmd.Execute()
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}
