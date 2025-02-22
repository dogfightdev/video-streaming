#!/bin/bash

# Input file (get from argument)
if [ -z "$1" ]; then
  echo "Usage: $0 <input_file.mp4>"
  exit 1
fi
INPUT_FILE="$1"

# Check if input file exists
if [ ! -f "$INPUT_FILE" ]; then
  echo "Error: Input file '$INPUT_FILE' not found."
  exit 1
fi

# Output directory
OUTPUT_DIR="hls_output"
mkdir -p "$OUTPUT_DIR"

# Video resolutions and bitrates (adjust as needed)
RESOLUTIONS=("1920x1080" "1280x720" "854x480" "640x360")
VIDEO_BITRATES=("5M" "3M" "1M" "500k")

# Audio bitrates (AAC)
AUDIO_BITRATES=("192k" "128k" "64k")

# Audio channel layouts
AUDIO_CHANNELS=("2")

# Segment duration
SEGMENT_DURATION=10

# Frame rate
FRAME_RATE=23.976

# Main manifest file
MAIN_MANIFEST="$OUTPUT_DIR/main.m3u8"

# Function to encode video
encode_video() {
  local resolution="$1"
  local bitrate="$2"
  local output_dir="$OUTPUT_DIR/video_${resolution}_${bitrate}"
  mkdir -p "$output_dir"
  local output_manifest="$output_dir/prog_index.m3u8"
  local output_segment="$output_dir/seg_%04d.ts"
  local width=$(echo "$resolution" | cut -d'x' -f1)
  local height=$(echo "$resolution" | cut -d'x' -f2)

  ffmpeg -i "$INPUT_FILE" \
    -c:v hevc_videotoolbox \
    -b:v "$bitrate" \
    -s "${width}x${height}" \
    -r "$FRAME_RATE" \
    -hls_time "$SEGMENT_DURATION" \
    -hls_list_size 0 \
    -hls_segment_filename "$output_segment" \
    "$output_manifest"

  echo "#EXT-X-STREAM-INF:BANDWIDTH=$(get_bandwidth "$bitrate"),RESOLUTION=$resolution,CODECS=\"hvc1.2.4.L153\",AUDIO=\"audio\"" >> "$MAIN_MANIFEST"
  echo "${output_dir##*/}/prog_index.m3u8" >> "$MAIN_MANIFEST"

  # I-frame only generation
  local iframe_dir="$OUTPUT_DIR/iframe_${resolution}_${bitrate}"
  mkdir -p "$iframe_dir"
  local iframe_manifest="$iframe_dir/iframe_index.m3u8"
  local iframe_segment="$iframe_dir/iframe_seg_%04d.ts"

  ffmpeg -skip_frame nokey -i "$INPUT_FILE" \
    -vsync 0 \
    -c:v hevc_videotoolbox \
    -s "${width}x${height}" \
    -r "$FRAME_RATE" \
    -hls_time "$SEGMENT_DURATION" \
    -hls_list_size 0 \
    -hls_segment_filename "$iframe_segment" \
    "$iframe_manifest"

  echo "#EXT-X-I-FRAME-STREAM-INF:BANDWIDTH=$(get_bandwidth "$bitrate"),RESOLUTION=$resolution,URI=\"${iframe_dir##*/}/iframe_index.m3u8\",CODECS=\"hvc1.2.4.L153\"" >> "$MAIN_MANIFEST"

}

# Function to encode audio
encode_audio() {
  local bitrate="$1"
  local channels="$2"
  local output_dir="$OUTPUT_DIR/audio_${bitrate}_${channels}ch"
  mkdir -p "$output_dir"
  local output_manifest="$output_dir/prog_index.m3u8"
  local output_segment="$output_dir/seg_%04d.aac"

  ffmpeg -i "$INPUT_FILE" \
    -map 0:a:0 \
    -c:a aac \
    -b:a "$bitrate" \
    -ac "$channels" \
    -hls_time "$SEGMENT_DURATION" \
    -hls_list_size 0 \
    -hls_segment_filename "$output_segment" \
    "$output_manifest"

  echo "#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID=\"audio\",NAME=\"English\",LANGUAGE=\"en\",AUTOSELECT=YES,DEFAULT=YES,CHANNELS=\"$channels\",URI=\"${output_dir##*/}/prog_index.m3u8\"" >> "$MAIN_MANIFEST"

}

# Function to convert subtitles to WebVTT
encode_subtitles() {
  local output_dir="$OUTPUT_DIR/subtitles_en"
  mkdir -p "$output_dir"
  local output_manifest="$output_dir/prog_index.m3u8"
  local output_segment="$output_dir/seg_%04d.vtt"

  ffmpeg -i "$INPUT_FILE" \
    -map 0:s:0 \
    -c:s webvtt \
    -hls_time "$SEGMENT_DURATION" \
    -hls_list_size 0 \
    -hls_segment_filename "$output_segment" \
    "$output_manifest"

  echo "#EXT-X-MEDIA:TYPE=SUBTITLES,GROUP-ID=\"subtitles\",NAME=\"English\",LANGUAGE=\"en\",AUTOSELECT=YES,DEFAULT=YES,URI=\"${output_dir##*/}/prog_index.m3u8\"" >> "$MAIN_MANIFEST"
}

# Function to estimate bandwidth
get_bandwidth() {
  local bitrate="$1"
  local value=$(echo "$bitrate" | sed 's/[^0-9]*//g')
  local unit=$(echo "$bitrate" | sed 's/[0-9]*//g')
  if [[ "$unit" == "k" ]]; then
    echo $((value * 1000))
  elif [[ "$unit" == "M" ]]; then
    echo $((value * 1000000))
  else
    echo "$value"
  fi
}

# Main script
echo "#EXTM3U" > "$MAIN_MANIFEST"
echo "#EXT-X-VERSION:6" >> "$MAIN_MANIFEST"
echo "#EXT-X-INDEPENDENT-SEGMENTS" >> "$MAIN_MANIFEST"

# Audio encoding
for bitrate in "${AUDIO_BITRATES[@]}"; do
  for channels in "${AUDIO_CHANNELS[@]}"; do
    encode_audio "$bitrate" "$channels"
  done
done

# Subtitle encoding
encode_subtitles

# Video encoding
for resolution in "${RESOLUTIONS[@]}"; do
  for bitrate in "${VIDEO_BITRATES[@]}"; do
    encode_video "$resolution" "$bitrate"
  done
done

echo "#EXT-X-MEDIA:TYPE=SUBTITLES,GROUP-ID=\"subtitles\",NAME=\"English\",LANGUAGE=\"en\",AUTOSELECT=YES,DEFAULT=YES,URI=\"subtitles_en/prog_index.m3u8\"" >> "$MAIN_MANIFEST"

echo "HLS manifest generated in $OUTPUT_DIR"