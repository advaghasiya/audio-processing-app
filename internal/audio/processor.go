package audio

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/youpy/go-wav"
)

type AudioInfo struct {
	OriginalSampleRate  int
	ResampledSampleRate int
	Duration            float64
	Channels            int
	BitsPerSample       int
	ProcessingTime      float64
	IntegrityMessage    string
	DurationMessage     string
	Title               string
	Artist              string
	Album               string
	Genre               string
	Year                int
	Bitrate             int
}

func ProcessAudio(filePath string, targetSampleRate int, allowedSampleRates map[int]bool) (*AudioInfo, error) {
	startTime := time.Now()

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := wav.NewReader(file)
	format, err := reader.Format()
	if err != nil {
		return nil, fmt.Errorf("failed to read WAV format: %w", err)
	}

	if !allowedSampleRates[int(format.SampleRate)] {
		return nil, errors.New("unsupported sample rate")
	}

	duration, err := reader.Duration()
	if err != nil {
		return nil, fmt.Errorf("failed to get duration: %w", err)
	}

	if duration.Seconds() < 0.1 || duration.Seconds() > 600 {
		return nil, errors.New("audio duration is out of accepted range (0.1-600 seconds)")
	}

	// In a real-world scenario, you would implement resampling here
	// For simplicity, we're just checking if resampling is needed
	resampledSampleRate := int(format.SampleRate)
	if int(format.SampleRate) != targetSampleRate {
		resampledSampleRate = targetSampleRate
		// Implement resampling logic here
	}

	info := &AudioInfo{
		OriginalSampleRate:  int(format.SampleRate),
		ResampledSampleRate: resampledSampleRate,
		Duration:            duration.Seconds(),
		Channels:            int(format.NumChannels),
		BitsPerSample:       int(format.BitsPerSample),
		ProcessingTime:      time.Since(startTime).Seconds(),
		IntegrityMessage:    "WAV file is valid",
		DurationMessage:     fmt.Sprintf("Audio duration: %.2f seconds", duration.Seconds()),
	}

	// In a real-world scenario, you would extract metadata here
	// For simplicity, we're leaving these fields empty
	info.Title = ""
	info.Artist = ""
	info.Album = ""
	info.Genre = ""
	info.Year = 0
	info.Bitrate = int(format.SampleRate) * int(format.BitsPerSample) * int(format.NumChannels)

	return info, nil
}
