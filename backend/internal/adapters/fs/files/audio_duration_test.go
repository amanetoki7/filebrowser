package files

import (
	"context"
	"encoding/binary"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/gtsteffaniak/filebrowser/backend/internal/ffmpeg"
	"github.com/gtsteffaniak/filebrowser/backend/pkg/indexing/iteminfo"
)

func TestAudioDurationIndependentOfTagSupportAndSize(t *testing.T) {
	exe, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not installed")
	}
	if err := ffmpeg.Initialize(context.Background(), ffmpeg.InitOptions{FFmpegPath: exe, CacheDir: t.TempDir(), SkipHWTests: true, MaxConcurrent: 2}); err != nil {
		t.Fatal(err)
	}
	for _, sample := range []struct {
		name      string
		dataBytes uint32
		want      int
	}{{"tagless", 32000, 2}, {"subsecond", 4000, 1}, {"above_tag_size_limit", 301 * 1024 * 1024, 301 * 1024 * 1024 / 16000}} {
		t.Run(sample.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), sample.name+".wav")
			header := make([]byte, 44)
			copy(header, "RIFF")
			binary.LittleEndian.PutUint32(header[4:], sample.dataBytes+36)
			copy(header[8:], "WAVEfmt ")
			binary.LittleEndian.PutUint32(header[16:], 16)
			binary.LittleEndian.PutUint16(header[20:], 1)
			binary.LittleEndian.PutUint16(header[22:], 1)
			binary.LittleEndian.PutUint32(header[24:], 8000)
			binary.LittleEndian.PutUint32(header[28:], 16000)
			binary.LittleEndian.PutUint16(header[32:], 2)
			binary.LittleEndian.PutUint16(header[34:], 16)
			copy(header[36:], "data")
			binary.LittleEndian.PutUint32(header[40:], sample.dataBytes)
			if err := os.WriteFile(path, header, 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.Truncate(path, int64(sample.dataBytes)+44); err != nil {
				t.Fatal(err)
			}
			item := &iteminfo.ExtendedItemInfo{}
			if err := extractAudioMetadata(context.Background(), item, path, false, true, ffmpeg.Get()); err != nil {
				t.Fatal(err)
			}
			if item.Metadata == nil || item.Metadata.Duration != sample.want {
				t.Fatalf("got %+v; want %ds despite unreadable tags", item.Metadata, sample.want)
			}
		})
	}
}
