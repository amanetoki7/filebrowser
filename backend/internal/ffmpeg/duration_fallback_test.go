package ffmpeg

import (
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestProbeDurationFallbackFields(t *testing.T) {
	cases := []struct {
		name, input string
		want        float64
	}{
		{"container", `{"format":{"duration":"12.5"},"streams":[{"codec_type":"audio","duration":"99"}]}`, 12.5},
		{"longest media stream", `{"format":{"duration":"N/A"},"streams":[{"codec_type":"audio","duration":"12.5"},{"codec_type":"video","duration":"13"},{"codec_type":"subtitle","duration":"999"}]}`, 13},
		{"timestamp timebase", `{"streams":[{"codec_type":"audio","duration_ts":96000,"time_base":"1/48000"}]}`, 2},
		{"duration tag", `{"streams":[{"codec_type":"audio","tags":{"DURATION":"01:02:03.5"}}]}`, 3723.5},
		{"unavailable", `{"format":{"duration":"N/A"},"streams":[{"codec_type":"audio","duration":"0"}]}`, 0},
		{"nonfinite", `{"format":{"duration":"NaN"},"streams":[{"codec_type":"audio","duration":"Infinity"}]}`, 0},
		{"negative", `{"streams":[{"codec_type":"audio","duration":"-2","duration_ts":-4,"time_base":"1/2"}]}`, 0},
		{"invalid tag", `{"streams":[{"codec_type":"audio","tags":{"DURATION":"00:99:00"}}]}`, 0},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseProbeDuration([]byte(tt.input))
			if tt.want == 0 {
				if err == nil {
					t.Fatalf("expected unavailable, got %v", got)
				}
				return
			}
			if err != nil || math.Abs(got-tt.want) > 0.0001 {
				t.Fatalf("got %v, %v; want %v", got, err, tt.want)
			}
		})
	}
}

func TestPCMHeaderDurationWithoutTags(t *testing.T) {
	const rate = 8000
	data := make([]byte, 44+rate*2*2)
	copy(data, "RIFF")
	binary.LittleEndian.PutUint32(data[4:], uint32(len(data)-8))
	copy(data[8:], "WAVEfmt ")
	binary.LittleEndian.PutUint32(data[16:], 16)
	binary.LittleEndian.PutUint16(data[20:], 1)
	binary.LittleEndian.PutUint16(data[22:], 1)
	binary.LittleEndian.PutUint32(data[24:], rate)
	binary.LittleEndian.PutUint32(data[28:], rate*2)
	binary.LittleEndian.PutUint16(data[32:], 2)
	binary.LittleEndian.PutUint16(data[34:], 16)
	copy(data[36:], "data")
	binary.LittleEndian.PutUint32(data[40:], uint32(len(data)-44))
	path := filepath.Join(t.TempDir(), "tagless.wav")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	got, err := readPCMDuration(path)
	if err != nil || got != 2 {
		t.Fatalf("got %v, %v; want 2s", got, err)
	}
	if err := os.Truncate(path, 45); err != nil {
		t.Fatal(err)
	}
	if _, err := readPCMDuration(path); err == nil {
		t.Fatal("truncated data must not produce an invented duration")
	}
}
