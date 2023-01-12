package main

import (
	"encoding/json"
	"fmt"
	"github.com/snapcore/snapd/asserts"
	"github.com/snapcore/snapd/syncloud"
	"github.com/spf13/cobra"
	"strconv"
)

func main() {

	var rootCmd = &cobra.Command{Use: "syncloud-release"}

	var target string
	rootCmd.PersistentFlags().StringVarP(&target, "target", "t", "s3", "target: s3 or local dir")

	var file string
	var branch string
	var storage Storage
	var cmdPublish = &cobra.Command{
		Use:   "publish",
		Short: "Publish an app to Syncloud Store",
		Args:  cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			sha384, size, err := asserts.SnapFileSHA3_384(file)
			Check(err)
			info, err := Parse(file, branch)
			Check(err)
			storage = NewStorage(target)
			Check(storage.UploadFile(file, info.StoreSnapPath))
			Check(storage.UploadContent(sha384, info.StoreSha384Path))
			sizeString := strconv.FormatUint(size, 10)
			Check(storage.UploadContent(sizeString, info.StoreSizePath))
			Check(storage.UploadContent(info.Version, info.StoreVersionPath))
			snapRevision := &syncloud.SnapRevision{
				Id:       syncloud.ConstructSnapId(info.Name, info.Version),
				Size:     sizeString,
				Revision: info.Version,
				Sha384:   sha384,
			}
			snapRevisionJson, err := json.Marshal(snapRevision)
			Check(err)
			Check(storage.UploadContent(string(snapRevisionJson), fmt.Sprintf("revisions/%s.revision", sha384)))

		},
	}
	cmdPublish.Flags().StringVarP(&file, "file", "f", "", "snap file path")
	Check(cmdPublish.MarkFlagRequired("file"))
	cmdPublish.Flags().StringVarP(&branch, "branch", "b", "", "branch")
	Check(cmdPublish.MarkFlagRequired("branch"))
	rootCmd.AddCommand(cmdPublish)

	var app string
	var arch string
	var cmdPromote = &cobra.Command{
		Use:   "promote",
		Short: "Promote an app to stable channel",
		Args:  cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			storage = NewStorage(target)
			version := storage.DownloadContent(fmt.Sprintf("releases/rc/%s.%s.version", app, arch))
			Check(storage.UploadContent(version, fmt.Sprintf("releases/stable/%s.%s.version", app, arch)))
		},
	}
	cmdPromote.Flags().StringVarP(&app, "name", "n", "", "app name to promote")
	Check(cmdPromote.MarkFlagRequired("name"))
	cmdPromote.Flags().StringVarP(&arch, "arch", "a", "", "arch to promote")
	Check(cmdPromote.MarkFlagRequired("arch"))
	rootCmd.AddCommand(cmdPromote)

	var channel string
	var version string
	var cmdSetVersion = &cobra.Command{
		Use:   "set-version",
		Short: "Set app version on a channel",
		Args:  cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			storage = NewStorage(target)
			Check(storage.UploadContent(version, fmt.Sprintf("releases/%s/%s.%s.version", channel, app, arch)))
		},
	}
	cmdSetVersion.Flags().StringVarP(&app, "name", "n", "", "app")
	Check(cmdSetVersion.MarkFlagRequired("name"))
	cmdSetVersion.Flags().StringVarP(&arch, "arch", "a", "", "arch")
	Check(cmdSetVersion.MarkFlagRequired("arch"))
	cmdSetVersion.Flags().StringVarP(&version, "version", "v", "", "version")
	Check(cmdSetVersion.MarkFlagRequired("version"))
	cmdSetVersion.Flags().StringVarP(&channel, "channel", "c", "", "channel")
	Check(cmdSetVersion.MarkFlagRequired("channel"))
	rootCmd.AddCommand(cmdSetVersion)

	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}

func NewStorage(target string) Storage {
	if target == "s3" {
		return NewS3("apps.syncloud.org")
	} else {
		return NewFileSystem(target)
	}
}

func Check(err error) {
	if err != nil {
		panic(err)
	}
}
