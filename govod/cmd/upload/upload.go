package upload

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"net/http"
	"time"

	"github.com/spf13/cobra"
)

var (
	packagePath  string
	rcloneRemote string
	contentAPI   string
)

var UploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload HLS package and send metadata",
	Run: func(cmd *cobra.Command, args []string) {
		if packagePath == "" || rcloneRemote == "" || contentAPI == "" {
			fmt.Println("Error: --package, --rclone-remote, and --content-api are required.")
			cmd.Usage()
			os.Exit(1)
		}

		// 1. Upload using rclone
		if err := uploadWithRclone(packagePath, rcloneRemote); err != nil {
			log.Fatalf("Error uploading with rclone: %v", err)
		}

		// 2. Extract metadata using ffmpeg
		metadata, err := extractMetadata(filepath.Join(packagePath, "main.m3u8"))
		if err != nil {
			log.Fatalf("Error extracting metadata: %v", err)
		}

		// 3. Send metadata to content API
		if err := sendMetadata(metadata, contentAPI); err != nil {
			log.Fatalf("Error sending metadata: %v", err)
		}

		fmt.Println("Upload and metadata submission successful.")
	},
}

func init() {
	UploadCmd.Flags().StringVar(&packagePath, "package", "", "Path to the HLS package")
	UploadCmd.Flags().StringVar(&rcloneRemote, "rclone-remote", "", "Rclone remote configuration (e.g., r2:video-streaming-storage)")
	UploadCmd.Flags().StringVar(&contentAPI, "content-api", "", "Content API URL (e.g., https://api.dogfight.dev)")
}

func uploadWithRclone(packagePath, rcloneRemote string) error {
	cmd := exec.Command("rclone", "copy", packagePath, rcloneRemote, "-P")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func extractMetadata(manifestPath string) (map[string]interface{}, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_format",
		"-show_streams",
		"-of", "json",
		manifestPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ffprobe error: %s, output: %s", err, string(output))
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(output, &metadata); err != nil {
		return nil, err
	}

	return metadata, nil
}

func sendMetadata(metadata map[string]interface{}, contentAPI string) error {
	jsonData, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("POST", contentAPI, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned status: %s", resp.Status)
	}

	return nil
}
