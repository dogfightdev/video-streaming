package main

import (
	"github.com/dogfightdev/video-streaming/cmd/hls"
	"github.com/dogfightdev/video-streaming/cmd/upload"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{Use: "govod"}
	rootCmd.AddCommand(hls.HlsCmd)
	rootCmd.AddCommand(upload.UploadCmd)
	rootCmd.Execute()
}
