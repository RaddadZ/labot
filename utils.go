package main

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/exp/constraints"
	tele "gopkg.in/telebot.v3"
)

func min[T constraints.Ordered](x, y T) T {
	if x < y {
		return x
	}
	return y
}

func max[T constraints.Ordered](x, y T) T {
	if x > y {
		return x
	}
	return y
}

func abs[T constraints.Signed](x T) T {
	if x < 0 {
		return -1 * x
	}
	return x
}

func getFormattedDateTime(timestamp int64, timezoneHours int) string {
	date := time.Unix(timestamp, 0)
	date = date.UTC().Add(time.Hour * time.Duration(timezoneHours))
	return date.Format("Mon 15:04:05")
}

func getUserString(user *tele.User) string {
	if user.Username != "" {
		return user.Username
	}

	if user.FirstName != "" || user.LastName != "" {
		return strings.Join([]string{user.FirstName, user.LastName}, " ")
	}

	return fmt.Sprintf("%d", user.ID)
}

func formatTimePeriod[T int | int16 | int32 | int64](seconds T) string {
	hours := seconds / T(3600)
	minutes := (seconds % 3600) / 60
	seconds = seconds % 60
	if hours != 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func formatTimezone(hours int) string {
	sign := "+"
	if hours < 0 {
		sign = "-"
	}

	return fmt.Sprintf("%s%02d:00", sign, abs(hours))
}
