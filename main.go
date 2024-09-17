package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	tele "gopkg.in/telebot.v3"
)

const defaultLatestTimesCount = 5
const defaultTimezoneHour = 3
const clockEmoji = "\xF0\x9F\x95\x90"

var chatsTimes map[int64][]int64
var chatsTimeZones map[int64]int

var asd []int
var asdwd [4]int
var mapss map[string]int

func main() {
	chatsTimes = map[int64][]int64{}
	chatsTimeZones = map[int64]int{}

	pref := tele.Settings{
		Token:  os.Getenv("LABOT_TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	commands := []tele.Command{
		{Text: "/start", Description: "Start the bot."},
		{Text: "/timezone", Description: "Show currently used timezone hour."},
		{Text: "/alllog", Description: "Show all time log for chat."},
	}

	err = b.SetCommands(commands)
	if err != nil {
		log.Fatal(err)
		return
	}

	menu := &tele.ReplyMarkup{ResizeKeyboard: true}

	btnLogTime := menu.Text("Log Time")
	btnLatest10 := menu.Text("Latest 10 Logs")
	btnLast15Min := menu.Text("Last 15 Minutes")
	btnLastHour := menu.Text("Last Hour")

	menu.Reply(menu.Row(btnLogTime), menu.Row(btnLatest10, btnLast15Min, btnLastHour))

	b.Handle("/start", func(c tele.Context) error {
		user := getUserString(c.Sender())

		chatsTimeZones[c.Chat().ID] = defaultTimezoneHour

		return c.Send(fmt.Sprintf("Hello %s!\nTo keep track of your labor contractions, press the 'Log Time' button whenever you have a contraction.\nHave a safe birth!", user), menu)
	})

	selector := &tele.ReplyMarkup{}

	timezoneBtns := []tele.Btn{
		selector.Data("-3h", "minus-3", "-3"),
		selector.Data("-2h", "minus-2", "-2"),
		selector.Data("-1h", "minus-1", "-1"),
		selector.Data("+1h", "plus-1", "1"),
		selector.Data("+2h", "plus-2", "2"),
		selector.Data("+3h", "plus-3", "3"),
	}

	for _, btn := range timezoneBtns {
		b.Handle(&btn, func(c tele.Context) error {
			hours, err := strconv.Atoi(btn.Data)
			if err != nil {
				return nil
			}

			timezoneHours, ok := chatsTimeZones[c.Chat().ID]
			if !ok {
				timezoneHours = defaultTimezoneHour
			}

			newTimezone := timezoneHours + hours
			if newTimezone > 12 {
				newTimezone -= 24
			} else if newTimezone < -11 {
				newTimezone += 24
			}

			chatsTimeZones[c.Chat().ID] = newTimezone

			msgStr := fmt.Sprintf("Timezone successfully changed from %s to %s.", formatTimezone(timezoneHours), formatTimezone(newTimezone))

			c.Send(msgStr)
			return c.Respond()
		})
	}

	selector.Inline(selector.Row(timezoneBtns...))

	b.Handle("/timezone", func(c tele.Context) error {
		timezoneHours, ok := chatsTimeZones[c.Chat().ID]
		if !ok {
			timezoneHours = defaultTimezoneHour
		}

		msgStr := fmt.Sprintf("Current timezone is %s.", formatTimezone(timezoneHours))

		return c.Send(msgStr, selector)
	})

	b.Handle(&btnLogTime, func(c tele.Context) error {
		messageTime := c.Message().Unixtime
		chatId := c.Chat().ID

		userData, ok := chatsTimes[chatId]
		if !ok {
			userData = []int64{}
		}

		userData = append(userData, messageTime)
		chatsTimes[chatId] = userData

		latestTimesDiffSum := 0
		latestTimesCount := min(defaultLatestTimesCount, len(userData))
		latestTimes := userData[len(userData)-latestTimesCount:]

		timezoneHours, ok := chatsTimeZones[c.Chat().ID]
		if !ok {
			timezoneHours = defaultTimezoneHour
		}

		latestTimesStr := fmt.Sprintf("%s %s\n", clockEmoji, getFormattedDateTime(latestTimes[0], timezoneHours))

		for i := 1; i < latestTimesCount; i++ {
			diff := int(latestTimes[i] - latestTimes[i-1])
			latestTimesDiffSum += diff
			latestTimesStr += fmt.Sprintf("%s %s  (+%s)\n", clockEmoji, getFormattedDateTime(latestTimes[i], timezoneHours), formatTimePeriod(diff))
		}

		latestTimesAvg := float64(latestTimesDiffSum) / float64(max(latestTimesCount-1, 1))

		msgStr := latestTimesStr + fmt.Sprintf("\nlatest %d times average: %s.", latestTimesCount, formatTimePeriod(int(latestTimesAvg)))

		return c.Send(msgStr)
	})

	b.Handle("/alllog", func(c tele.Context) error {
		chatId := c.Chat().ID

		userData, ok := chatsTimes[chatId]
		if !ok || len(userData) == 0 {
			c.Send("No log.")
			return nil
		}

		latestTimesDiffSum := 0
		latestTimesCount := len(userData)
		latestTimes := userData

		timezoneHours, ok := chatsTimeZones[c.Chat().ID]
		if !ok {
			timezoneHours = defaultTimezoneHour
		}

		latestTimesStr := fmt.Sprintf("%s %s\n", clockEmoji, getFormattedDateTime(latestTimes[0], timezoneHours))

		for i := 1; i < latestTimesCount; i++ {
			diff := int(latestTimes[i] - latestTimes[i-1])
			latestTimesDiffSum += diff
			latestTimesStr += fmt.Sprintf("%s %s  (+%s)\n", clockEmoji, getFormattedDateTime(latestTimes[i], timezoneHours), formatTimePeriod(diff))
		}

		latestTimesAvg := float64(latestTimesDiffSum) / float64(max(latestTimesCount-1, 1))

		msgStr := latestTimesStr + fmt.Sprintf("\nlatest %d times average: %s.", latestTimesCount, formatTimePeriod(int(latestTimesAvg)))

		return c.Send(msgStr)
	})

	b.Handle(&btnLatest10, func(c tele.Context) error {
		chatId := c.Chat().ID

		userData, ok := chatsTimes[chatId]
		if !ok || len(userData) == 0 {
			c.Send("No log.")
			return nil
		}

		TenLatestTimesCount := 10

		latestTimesDiffSum := 0
		latestTimesCount := min(TenLatestTimesCount, len(userData))
		latestTimes := userData[len(userData)-latestTimesCount:]

		timezoneHours, ok := chatsTimeZones[c.Chat().ID]
		if !ok {
			timezoneHours = defaultTimezoneHour
		}

		latestTimesStr := fmt.Sprintf("%s %s\n", clockEmoji, getFormattedDateTime(latestTimes[0], timezoneHours))

		for i := 1; i < latestTimesCount; i++ {
			diff := int(latestTimes[i] - latestTimes[i-1])
			latestTimesDiffSum += diff
			latestTimesStr += fmt.Sprintf("%s %s  (+%s)\n", clockEmoji, getFormattedDateTime(latestTimes[i], timezoneHours), formatTimePeriod(diff))
		}

		latestTimesAvg := float64(latestTimesDiffSum) / float64(max(latestTimesCount-1, 1))

		msgStr := latestTimesStr + fmt.Sprintf("\nlatest %d times average: %s.", latestTimesCount, formatTimePeriod(int(latestTimesAvg)))

		return c.Send(msgStr)
	})

	b.Handle(&btnLast15Min, func(c tele.Context) error {
		messageTime := c.Message().Unixtime
		chatId := c.Chat().ID

		userData, ok := chatsTimes[chatId]
		if !ok || len(userData) == 0 {
			c.Send("No log.")
			return nil
		}

		logsAfter := messageTime - 15*60
		latestTimes := []int64{}
		for _, v := range userData {
			if v > logsAfter {
				latestTimes = append(latestTimes, v)
			}
		}

		if len(latestTimes) == 0 {
			c.Send("No log.")
			return nil
		}

		latestTimesDiffSum := 0
		latestTimesCount := len(latestTimes)

		timezoneHours, ok := chatsTimeZones[c.Chat().ID]
		if !ok {
			timezoneHours = defaultTimezoneHour
		}

		latestTimesStr := fmt.Sprintf("%s %s\n", clockEmoji, getFormattedDateTime(latestTimes[0], timezoneHours))

		for i := 1; i < latestTimesCount; i++ {
			diff := int(latestTimes[i] - latestTimes[i-1])
			latestTimesDiffSum += diff
			latestTimesStr += fmt.Sprintf("%s %s  (+%s)\n", clockEmoji, getFormattedDateTime(latestTimes[i], timezoneHours), formatTimePeriod(diff))
		}

		latestTimesAvg := float64(latestTimesDiffSum) / float64(max(latestTimesCount-1, 1))

		msgStr := latestTimesStr + fmt.Sprintf("\nlatest %d times average: %s.", latestTimesCount, formatTimePeriod(int(latestTimesAvg)))

		return c.Send(msgStr)
	})

	b.Handle(&btnLastHour, func(c tele.Context) error {
		messageTime := c.Message().Unixtime
		chatId := c.Chat().ID

		userData, ok := chatsTimes[chatId]
		if !ok || len(userData) == 0 {
			c.Send("No log.")
			return nil
		}

		logsAfter := messageTime - 60*60
		latestTimes := []int64{}
		for _, v := range userData {
			if v > logsAfter {
				latestTimes = append(latestTimes, v)
			}
		}

		if len(latestTimes) == 0 {
			c.Send("No log.")
			return nil
		}

		latestTimesDiffSum := 0
		latestTimesCount := len(latestTimes)

		timezoneHours, ok := chatsTimeZones[c.Chat().ID]
		if !ok {
			timezoneHours = defaultTimezoneHour
		}

		latestTimesStr := fmt.Sprintf("%s %s\n", clockEmoji, getFormattedDateTime(latestTimes[0], timezoneHours))

		for i := 1; i < latestTimesCount; i++ {
			diff := int(latestTimes[i] - latestTimes[i-1])
			latestTimesDiffSum += diff
			latestTimesStr += fmt.Sprintf("%s %s  (+%s)\n", clockEmoji, getFormattedDateTime(latestTimes[i], timezoneHours), formatTimePeriod(diff))
		}

		latestTimesAvg := float64(latestTimesDiffSum) / float64(max(latestTimesCount-1, 1))

		msgStr := latestTimesStr + fmt.Sprintf("\nlatest %d times average: %s.", latestTimesCount, formatTimePeriod(int(latestTimesAvg)))

		return c.Send(msgStr)
	})

	b.Start()
}
