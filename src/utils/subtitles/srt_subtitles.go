package subtitles

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/pluja/anysub/models"
)

func GenerateSubsSRT(t models.TranscriptionResult, c SubtitleConfig) string {
	var srtBuilder strings.Builder

	speakerColors := make(map[string]string)
	if c.ColorizeSpeakers {
		speakerColors = srtSpeakerColors(t, c.SpeakerColors)
	}

	buildWordCues(t, c, func(cueCounter int, startTime, endTime float64, speaker string, words []string) {
		writeSrtCue(&srtBuilder, cueCounter, startTime, endTime, speakerColors[speaker], words)
	})

	return srtBuilder.String()
}

func writeSrtCue(b *strings.Builder, cueCounter int, startTime, endTime float64, color string, words []string) {
	b.WriteString(fmt.Sprintf("%d\n", cueCounter))
	b.WriteString(fmt.Sprintf("%s --> %s\n", formatSRTTime(startTime), formatSRTTime(endTime)))
	cueText := strings.Join(words, " ")
	if color != "" {
		// <font color> is the de-facto SRT styling tag (VLC, mpv, Jellyfin).
		cueText = fmt.Sprintf("<font color=\"%s\">%s</font>", color, cueText)
	}

	b.WriteString(cueText + "\n\n")
}

func srtSpeakerColors(t models.TranscriptionResult, customColors map[string]string) map[string]string {
	speakerColors := make(map[string]string)

	for _, segment := range t.Segments {
		for _, word := range segment.Words {
			if _, ok := speakerColors[word.Speaker]; !ok && word.Speaker != "" {
				if color, ok := customColors[word.Speaker]; ok {
					speakerColors[word.Speaker] = color
				} else {
					color, _ = slice.Random(vttColors)
					speakerColors[word.Speaker] = color
				}
			}
		}
	}
	return speakerColors
}

// formatSRTTime renders HH:MM:SS,mmm (SRT uses a comma before milliseconds).
func formatSRTTime(seconds float64) string {
	if seconds < 0 {
		seconds = 0
	}
	t := time.Duration(math.Round(seconds*1000)) * time.Millisecond
	hours := t / time.Hour
	t -= hours * time.Hour
	minutes := t / time.Minute
	t -= minutes * time.Minute
	secs := t / time.Second
	t -= secs * time.Second
	return fmt.Sprintf("%02d:%02d:%02d,%03d", hours, minutes, secs, t/time.Millisecond)
}
