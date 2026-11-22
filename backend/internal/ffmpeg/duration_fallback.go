package ffmpeg

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	goffmpeg "github.com/gtsteffaniak/go-ffmpeg"
)

// fallbackMediaDuration uses stream-level values rather than relying solely on
// the container duration. PCM WAV has a final, independent header-based path.
func (s *Service) fallbackMediaDuration(ctx context.Context, path string) (float64, error) {
	probeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := s.inner.Acquire(probeCtx, goffmpeg.SlotProbe); err == nil {
		cmd := exec.CommandContext(probeCtx, s.FFprobePath(), "-v", "error",
			"-protocol_whitelist", "file", "-show_entries",
			"format=duration:stream=codec_type,duration,duration_ts,time_base:stream_tags=DURATION",
			"-of", "json", path)
		hideDurationProbeWindow(cmd)
		output, probeErr := cmd.Output()
		s.inner.Release(goffmpeg.SlotProbe)
		if probeErr == nil {
			if duration, err := parseProbeDuration(output); err == nil {
				return duration, nil
			}
		}
	}
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}
	if duration, err := readPCMDuration(path); err == nil {
		return duration, nil
	}
	return 0, fmt.Errorf("media duration unavailable after container, stream and PCM-header probes")
}

func parseProbeDuration(data []byte) (float64, error) {
	var probe struct {
		Format struct {
			Duration json.RawMessage `json:"duration"`
		} `json:"format"`
		Streams []struct {
			Type     string            `json:"codec_type"`
			Duration json.RawMessage   `json:"duration"`
			Ticks    json.RawMessage   `json:"duration_ts"`
			TimeBase string            `json:"time_base"`
			Tags     map[string]string `json:"tags"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return 0, err
	}
	if duration := probeNumber(probe.Format.Duration); validMediaDuration(duration) {
		return duration, nil
	}
	var longest float64
	for _, stream := range probe.Streams {
		if stream.Type != "audio" && stream.Type != "video" {
			continue
		}
		duration := probeNumber(stream.Duration)
		if !validMediaDuration(duration) {
			parts := strings.Split(stream.TimeBase, "/")
			if len(parts) == 2 {
				numerator, _ := strconv.ParseFloat(parts[0], 64)
				denominator, _ := strconv.ParseFloat(parts[1], 64)
				ticks := probeNumber(stream.Ticks)
				if numerator > 0 && denominator > 0 && ticks > 0 {
					duration = ticks * numerator / denominator
				}
			}
		}
		if !validMediaDuration(duration) {
			for key, value := range stream.Tags {
				if strings.EqualFold(key, "DURATION") {
					duration = durationTag(value)
					break
				}
			}
		}
		if validMediaDuration(duration) && duration > longest {
			longest = duration
		}
	}
	if !validMediaDuration(longest) {
		return 0, fmt.Errorf("probe has no valid audio/video duration")
	}
	return longest, nil
}

func probeNumber(raw json.RawMessage) float64 {
	value, _ := strconv.ParseFloat(strings.Trim(string(raw), "\""), 64)
	return value
}

func durationTag(value string) float64 {
	parts := strings.Split(value, ":")
	if len(parts) != 3 {
		return 0
	}
	hours, errH := strconv.ParseFloat(parts[0], 64)
	minutes, errM := strconv.ParseFloat(parts[1], 64)
	seconds, errS := strconv.ParseFloat(parts[2], 64)
	if errH != nil || errM != nil || errS != nil || hours < 0 || minutes < 0 || minutes >= 60 || seconds < 0 || seconds >= 60 {
		return 0
	}
	return hours*3600 + minutes*60 + seconds
}

func readPCMDuration(path string) (float64, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return 0, err
	}
	var header [12]byte
	if _, err = io.ReadFull(file, header[:]); err != nil {
		return 0, err
	}
	if string(header[8:]) != "WAVE" {
		return 0, fmt.Errorf("not a WAV file")
	}
	var order binary.ByteOrder = binary.LittleEndian
	if string(header[:4]) == "RIFX" {
		order = binary.BigEndian
	} else if string(header[:4]) != "RIFF" {
		return 0, fmt.Errorf("unsupported WAV container")
	}
	var byteRate uint32
	var dataBytes int64
	for chunks := 0; chunks < 4096; chunks++ {
		var chunk [8]byte
		if _, err = io.ReadFull(file, chunk[:]); err != nil {
			break
		}
		size := int64(order.Uint32(chunk[4:]))
		start, _ := file.Seek(0, io.SeekCurrent)
		if size > info.Size()-start {
			return 0, fmt.Errorf("truncated WAV chunk")
		}
		switch string(chunk[:4]) {
		case "fmt ":
			if size < 16 {
				return 0, fmt.Errorf("invalid WAV format chunk")
			}
			var format [16]byte
			if _, err = io.ReadFull(file, format[:]); err != nil {
				return 0, err
			}
			codec := order.Uint16(format[:2])
			if codec != 1 && codec != 3 {
				return 0, fmt.Errorf("not PCM or IEEE-float WAV")
			}
			sampleRate := order.Uint32(format[4:8])
			byteRate = order.Uint32(format[8:12])
			blockAlign := order.Uint16(format[12:14])
			if uint64(byteRate) != uint64(sampleRate)*uint64(blockAlign) {
				return 0, fmt.Errorf("invalid PCM byte rate")
			}
		case "data":
			dataBytes = size
		}
		if byteRate > 0 && dataBytes > 0 {
			return float64(dataBytes) / float64(byteRate), nil
		}
		if _, err = file.Seek(start+size+(size%2), io.SeekStart); err != nil {
			return 0, err
		}
	}
	return 0, fmt.Errorf("WAV has no valid duration")
}
