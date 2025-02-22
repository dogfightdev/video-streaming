package hls

import (
	"fmt"
	"log"
	"os"

	"github.com/dogfightdev/video-streaming/encoder"
	"github.com/spf13/cobra"
)

var (
	inputFilePath   string
	outputDir       string
	segmentDuration int
	frameRate       float64
)

var HlsCmd = &cobra.Command{
	Use:   "hls",
	Short: "Generate optimized HLS manifest and segments",
	Run: func(cmd *cobra.Command, args []string) {
		if inputFilePath == "" {
			fmt.Println("Error: --input is required.")
			cmd.Usage()
			os.Exit(1)
		}

		if _, err := os.Stat(inputFilePath); os.IsNotExist(err) {
			log.Fatalf("Error: Input file '%s' not found.", inputFilePath)
		}

		if err := os.MkdirAll(outputDir, 0755); err != nil {
			log.Fatalf("Error creating output directory: %v, error: %v", outputDir, err)
		}

		// Encode Audio
		if err := encoder.EncodeAudio(inputFilePath, outputDir, "hls", segmentDuration); err != nil {
			log.Fatalf("Audio encoding failed: %v", err)
		}

		// Encode Video
		resolutions := []string{"1080p", "720p", "360p"}
		for _, resolution := range resolutions {
			if err := encoder.EncodeVideo(inputFilePath, outputDir, "hls", resolution, segmentDuration, frameRate); err != nil {
				log.Fatalf("Video encoding failed: %v", err)
			}
		}

		// Generate Master Playlist
		if err := encoder.GenerateMasterPlaylist(outputDir, "hls", resolutions); err != nil {
			log.Fatalf("Master playlist generation failed: %v", err)
		}

		fmt.Printf("HLS package successfully created at %s\n", outputDir)
	},
}

func init() {
	HlsCmd.Flags().StringVarP(&inputFilePath, "input", "i", "", "Input video file path")
	HlsCmd.Flags().StringVarP(&outputDir, "output", "o", "hls_output", "Output directory")
	HlsCmd.Flags().IntVarP(&segmentDuration, "segment-duration", "s", 10, "Segment duration in seconds")
	HlsCmd.Flags().Float64VarP(&frameRate, "frame-rate", "f", 23.976, "Frame rate")
}
