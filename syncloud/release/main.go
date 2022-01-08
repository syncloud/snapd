package main

import (
	"fmt"
	"github.com/snapcore/snapd/asserts"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"strconv"
)

func main() {

	var rootCmd = &cobra.Command{Use: "cli"}

	var file string
	var cmdPublish = &cobra.Command{
		Use:   "publish",
		Short: "Publish an app to Syncloud Store",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			sha384, size, err := asserts.SnapFileSHA3_384(file)
			if err != nil {
				log.Fatalf("error: %v\n", err)
			}
			err = ioutil.WriteFile(fmt.Sprintf("%s.sha384", file), []byte(sha384), 0644)
			if err != nil {
				log.Fatalf("error: %v\n", err)
			}
			err = ioutil.WriteFile(fmt.Sprintf("%s.size", file), []byte(strconv.FormatUint(size, 10)), 0644)
			if err != nil {
				log.Fatalf("error: %v\n", err)
			}
		},
	}
	cmdPublish.Flags().StringVarP(&file, "file", "f", "", "snap file path")
	rootCmd.AddCommand(cmdPublish)

	err := rootCmd.Execute()
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

}
