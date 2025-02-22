package manifest

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dogfightdev/video-streaming/utils"
)

type Manifest struct {
	mainManifest    string
	segmentDuration int
	frameRate       float64
	videoStreams    []VideoStream
	audioStreams    []AudioStream
	subtitleStream  SubtitleStream
}

type VideoStream struct {
	resolution string
	bitrate    string
}

type AudioStream struct {
	bitrate  string
	channels string
}

type SubtitleStream struct{}

func NewManifest(mainManifest string, segmentDuration int, frameRate float64) *Manifest {
	return &Manifest{
		mainManifest:    mainManifest,
		segmentDuration: segmentDuration,
		frameRate:       frameRate,
	}
}

func (m *Manifest) AddVideo(resolution, bitrate string) {
	m.videoStreams = append(m.videoStreams, VideoStream{
		resolution: resolution,
		bitrate:    bitrate,
	})
}

func (m *Manifest) AddAudio(bitrate, channels string) {
	m.audioStreams = append(m.audioStreams, AudioStream{
		bitrate:  bitrate,
		channels: channels,
	})
}

func (m *Manifest) AddSubtitles() {
	m.subtitleStream = SubtitleStream{}
}

func (m *Manifest) Write() error {
	file, err := os.Create(m.mainManifest)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString("#EXTM3U\n#EXT-X-VERSION:6\n#EXT-X-INDEPENDENT-SEGMENTS\n")
	if err != nil {
		return err
	}

	for _, audio := range m.audioStreams {
		audioDir := filepath.Join(filepath.Dir(m.mainManifest), fmt.Sprintf("audio_%s_%sch", audio.bitrate, audio.channels))
		_, err = file.WriteString(fmt.Sprintf("#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID=\"audio\",NAME=\"English\",LANGUAGE=\"en\",AUTOSELECT=YES,DEFAULT=YES,CHANNELS=\"%s\",URI=\"%s/prog_index.m3u8\"\n", audio.channels, filepath.Base(audioDir)))
		if err != nil {
			return err
		}
	}

	subtitleDir := filepath.Join(filepath.Dir(m.mainManifest), "subtitles_en")
	_, err = file.WriteString(fmt.Sprintf("#EXT-X-MEDIA:TYPE=SUBTITLES,GROUP-ID=\"subtitles\",NAME=\"English\",LANGUAGE=\"en\",AUTOSELECT=YES,DEFAULT=YES,URI=\"%s/prog_index.m3u8\"\n", filepath.Base(subtitleDir)))
	if err != nil {
		return err
	}

	for _, video := range m.videoStreams {
		videoDir := filepath.Join(filepath.Dir(m.mainManifest), fmt.Sprintf("video_%s_%s", video.resolution, video.bitrate))
		iframeDir := filepath.Join(filepath.Dir(m.mainManifest), fmt.Sprintf("iframe_%s_%s", video.resolution, video.bitrate))
		bandwidth := utils.GetBandwidth(video.bitrate)

		_, err = file.WriteString(fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%s,CODECS=\"hvc1.2.4.L153\",AUDIO=\"audio\",SUBTITLES=\"subtitles\"\n", bandwidth, video.resolution, filepath.Base(videoDir)))
		if err != nil {
			return err
		}
		_, err = file.WriteString(fmt.Sprintf("%s/prog_index.m3u8\n", filepath.Base(videoDir)))
		if err != nil {
			return err
		}

		_, err = file.WriteString(fmt.Sprintf("#EXT-X-I-FRAME-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%s,URI=\"%s/iframe_index.m3u8\",CODECS=\"hvc1.2.4.L153\"\n", bandwidth, video.resolution, filepath.Base(iframeDir)))
		if err != nil {
			return err
		}

	}

	return nil
}
