package handlers

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

func parseDate(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, errors.New("날짜를 입력해 주세요.")
	}
	return time.ParseInLocation("2006-01-02", value, time.Local)
}

func parseDateTime(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, errors.New("날짜/시간을 입력해 주세요.")
	}
	return time.ParseInLocation("2006-01-02T15:04", value, time.Local)
}

func parseOptionalDate(value string) (*time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	parsed, err := time.ParseInLocation("2006-01-02", value, time.Local)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseOptionalDateTime(value string) (*time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	parsed, err := time.ParseInLocation("2006-01-02T15:04", value, time.Local)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseOptionalInt64(value string) (*int64, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return nil, errors.New("숫자를 입력해 주세요.")
	}
	return &parsed, nil
}
