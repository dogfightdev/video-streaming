package encoder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func EncodeVideo(inputFilePath, outputDir, videoID, resolution string, segmentDuration int, frameRate float64) error {
	// Validate input file
	if _, err := os.Stat(inputFilePath); os.IsNotExist(err) {
		return fmt.Errorf("input file '%s' not found", inputFilePath)
	}

	outputVideoDir := filepath.Join(outputDir, videoID, "video", resolution)
	if err := os.MkdirAll(outputVideoDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	width, _ := strconv.Atoi(strings.TrimSuffix(resolution, "p"))
	height := (width * 9) / 16 // Ensure 16:9 aspect ratio

	bitrate, maxrate, bufsize := getBitrates(resolution)

	outputPlaylist := filepath.Join(outputVideoDir, "index.m3u8")
	outputSegment := filepath.Join(outputVideoDir, "seg_%04d.ts")

	// GOP size (must be a multiple of frame rate for better seeking)
	gopSize := int(frameRate * float64(segmentDuration))

	cmd := exec.Command("ffmpeg",
		"-i", inputFilePath,
		"-c:v", "hevc_videotoolbox", // ✅ Apple hardware-accelerated HEVC
		"-b:v", bitrate,
		"-maxrate", maxrate,
		"-bufsize", bufsize,
		"-s", fmt.Sprintf("%dx%d", width, height),
		"-r", fmt.Sprintf("%.3f", frameRate),
		"-g", fmt.Sprintf("%d", gopSize),
		"-keyint_min", fmt.Sprintf("%d", gopSize),
		"-pix_fmt", "nv12", // ✅ Correct pixel format for `hevc_videotoolbox`
		"-allow_sw", "1", // ✅ Allows software fallback if hardware fails
		"-hls_time", fmt.Sprintf("%d", segmentDuration),
		"-hls_list_size", "0",
		"-hls_segment_filename", outputSegment,
		"-hls_playlist_type", "vod",
		"-f", "hls", outputPlaylist,
	)

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("video encoding failed: %v", err)
	}

	fmt.Printf("✅ Video rendition %s generated at %s\n", resolution, outputVideoDir)
	return nil
}

func EncodeAudio(inputFilePath, outputDir, videoID string, segmentDuration int) error {
	outputAudioDir := filepath.Join(outputDir, videoID, "audio", "128k")
	if err := os.MkdirAll(outputAudioDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	outputPlaylist := filepath.Join(outputAudioDir, "audio.m3u8")
	outputSegment := filepath.Join(outputAudioDir, "seg_%04d.aac")

	cmd := exec.Command("ffmpeg",
		"-i", inputFilePath,
		"-map", "0:a:0",
		"-c:a", "aac", // ✅ Switched to FFmpeg's built-in AAC encoder
		"-b:a", "128k",
		"-ac", "2", // Stereo
		"-hls_time", fmt.Sprintf("%d", segmentDuration),
		"-hls_list_size", "0",
		"-hls_segment_filename", outputSegment,
		"-hls_playlist_type", "vod",
		"-f", "hls", outputPlaylist,
	)

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("audio encoding failed: %v", err)
	}

	fmt.Printf("✅ Audio encoded successfully at %s\n", outputAudioDir)
	return nil
}

func GenerateMasterPlaylist(outputDir, videoID string, resolutions []string) error {
	masterPlaylist := filepath.Join(outputDir, videoID, "master.m3u8")
	file, err := os.Create(masterPlaylist)
	if err != nil {
		return err
	}
	defer file.Close()

	file.WriteString("#EXTM3U\n")
	file.WriteString("#EXT-X-VERSION:3\n")

	for _, resolution := range resolutions {
		width, _ := strconv.Atoi(strings.TrimSuffix(resolution, "p"))
		bitrate, _, _ := getBitrates(resolution)
		file.WriteString(fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%s,RESOLUTION=%dx%d\n", bitrate, width, (width*9)/16))
		file.WriteString(fmt.Sprintf("video/%s/index.m3u8\n", resolution))
	}

	file.WriteString("#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID=\"audio\",NAME=\"English\",DEFAULT=YES,AUTOSELECT=YES,URI=\"audio/128k/audio.m3u8\"\n")

	return nil
}

func getBitrates(resolution string) (string, string, string) {
	switch resolution {
	case "1080p":
		return "5M", "6M", "10M"
	case "720p":
		return "3M", "4M", "8M"
	case "360p":
		return "1M", "2M", "4M"
	default:
		return "1M", "2M", "4M"
	}
}
