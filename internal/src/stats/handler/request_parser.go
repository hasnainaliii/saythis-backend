package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	statsdomain "saythis-backend/internal/src/stats/domain"
)

const dateLayout = "2006-01-02"

var allowedDailyPatchFields = map[string]struct{}{
	"date":               {},
	"mood":               {},
	"sleep_hours":        {},
	"journal_entry":      {},
	"stress_level":       {},
	"mindful_hours":      {},
	"stutter_score":      {},
	"stutter_count":      {},
	"repetition_count":   {},
	"filler_count":       {},
	"total_words":        {},
	"stutter_transcript": {},
}

func decodeDailyStatPatch(r *http.Request) (statsdomain.DailyStatPatch, error) {
	var payload map[string]json.RawMessage
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&payload); err != nil {
		return statsdomain.DailyStatPatch{}, fmt.Errorf("%w: %v", errInvalidRequestBody, err)
	}
	if len(payload) == 0 {
		return statsdomain.DailyStatPatch{}, errInvalidRequestBody
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return statsdomain.DailyStatPatch{}, errInvalidRequestBody
	}

	for field := range payload {
		if _, ok := allowedDailyPatchFields[field]; !ok {
			return statsdomain.DailyStatPatch{}, errInvalidRequestBody
		}
	}

	date, err := requiredDate(payload, "date")
	if err != nil {
		return statsdomain.DailyStatPatch{}, err
	}

	mood, err := optionalString(payload, "mood")
	if err != nil {
		return statsdomain.DailyStatPatch{}, err
	}
	sleepHours, err := optionalFloat(payload, "sleep_hours")
	if err != nil {
		return statsdomain.DailyStatPatch{}, err
	}
	journalEntry, err := optionalString(payload, "journal_entry")
	if err != nil {
		return statsdomain.DailyStatPatch{}, err
	}
	stressLevel, err := optionalInt(payload, "stress_level")
	if err != nil {
		return statsdomain.DailyStatPatch{}, err
	}
	mindfulHours, err := optionalFloat(payload, "mindful_hours")
	if err != nil {
		return statsdomain.DailyStatPatch{}, err
	}
	stutterScore, err := optionalFloat(payload, "stutter_score")
	if err != nil {
		return statsdomain.DailyStatPatch{}, err
	}
	stutterCount, err := optionalInt(payload, "stutter_count")
	if err != nil {
		return statsdomain.DailyStatPatch{}, err
	}
	repetitionCount, err := optionalInt(payload, "repetition_count")
	if err != nil {
		return statsdomain.DailyStatPatch{}, err
	}
	fillerCount, err := optionalInt(payload, "filler_count")
	if err != nil {
		return statsdomain.DailyStatPatch{}, err
	}
	totalWords, err := optionalInt(payload, "total_words")
	if err != nil {
		return statsdomain.DailyStatPatch{}, err
	}
	stutterTranscript, err := optionalString(payload, "stutter_transcript")
	if err != nil {
		return statsdomain.DailyStatPatch{}, err
	}

	return statsdomain.DailyStatPatch{
		Date:              date,
		Mood:              mood,
		SleepHours:        sleepHours,
		JournalEntry:      journalEntry,
		StressLevel:       stressLevel,
		MindfulHours:      mindfulHours,
		StutterScore:      stutterScore,
		StutterCount:      stutterCount,
		RepetitionCount:   repetitionCount,
		FillerCount:       fillerCount,
		TotalWords:        totalWords,
		StutterTranscript: stutterTranscript,
	}, nil
}

func requiredDate(payload map[string]json.RawMessage, field string) (time.Time, error) {
	raw, ok := payload[field]
	if !ok || isJSONNull(raw) {
		return time.Time{}, statsdomain.ErrDateRequired
	}

	var value string
	if err := json.Unmarshal(raw, &value); err != nil || value == "" {
		return time.Time{}, statsdomain.ErrInvalidDate
	}
	parsed, err := time.Parse(dateLayout, value)
	if err != nil {
		return time.Time{}, statsdomain.ErrInvalidDate
	}
	return parsed, nil
}

func optionalString(payload map[string]json.RawMessage, field string) (statsdomain.Optional[string], error) {
	raw, ok := payload[field]
	if !ok {
		return statsdomain.Optional[string]{}, nil
	}
	if isJSONNull(raw) {
		return statsdomain.Optional[string]{Present: true}, nil
	}

	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return statsdomain.Optional[string]{}, errInvalidRequestBody
	}
	return statsdomain.Optional[string]{Present: true, Value: &value}, nil
}

func optionalFloat(payload map[string]json.RawMessage, field string) (statsdomain.Optional[float64], error) {
	raw, ok := payload[field]
	if !ok {
		return statsdomain.Optional[float64]{}, nil
	}
	if isJSONNull(raw) {
		return statsdomain.Optional[float64]{Present: true}, nil
	}

	var value float64
	if err := json.Unmarshal(raw, &value); err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
		return statsdomain.Optional[float64]{}, errInvalidRequestBody
	}
	return statsdomain.Optional[float64]{Present: true, Value: &value}, nil
}

func optionalInt(payload map[string]json.RawMessage, field string) (statsdomain.Optional[int], error) {
	raw, ok := payload[field]
	if !ok {
		return statsdomain.Optional[int]{}, nil
	}
	if isJSONNull(raw) {
		return statsdomain.Optional[int]{Present: true}, nil
	}

	var value float64
	if err := json.Unmarshal(raw, &value); err != nil || math.Trunc(value) != value {
		return statsdomain.Optional[int]{}, errInvalidRequestBody
	}
	intValue := int(value)
	return statsdomain.Optional[int]{Present: true, Value: &intValue}, nil
}

func isJSONNull(raw json.RawMessage) bool {
	return bytes.Equal(bytes.TrimSpace(raw), []byte("null"))
}
