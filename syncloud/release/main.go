package main

import (
	"github.com/snapcore/snapd/asserts"
	"github.com/spf13/cobra"
	"strconv"
)

func main() {

	var rootCmd = &cobra.Command{Use: "cli"}

	var file string
	var branch string
	var target string
	var storage Storage
	var cmdPublish = &cobra.Command{
		Use:   "publish",
		Short: "Publish an app to Syncloud Store",
		Args:  cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			sha384, size, err := asserts.SnapFileSHA3_384(file)
			Check(err)
			info, err := Parse(file, branch)
			if target == "s3" {
				storage = NewS3("apps.syncloud.org")
			} else {
				storage = NewFileSystem(target)
			}
			Check(storage.UploadFile(file, info.StoreSnapPath))
			Check(err)
			Check(storage.UploadContent(sha384, info.StoreSha384Path))
			Check(storage.UploadContent(strconv.FormatUint(size, 10), info.StoreSizePath))
			Check(storage.UploadContent(info.Version, info.StoreVersionPath))
		},
	}
	cmdPublish.Flags().StringVarP(&file, "file", "f", "", "snap file path")
	Check(cmdPublish.MarkFlagRequired("file"))
	cmdPublish.Flags().StringVarP(&branch, "branch", "b", "", "branch")
	Check(cmdPublish.MarkFlagRequired("branch"))
	cmdPublish.Flags().StringVarP(&target, "target", "t", "s3", "target: s3 or local dir")
	rootCmd.AddCommand(cmdPublish)

	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}

func Check(err error) {
	if err != nil {
		panic(err)
	}
}
