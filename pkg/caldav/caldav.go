//  Vikunja is a todo-list application to facilitate your life.
//  Copyright 2018 Vikunja and contributors. All rights reserved.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with this program.  If not, see <https://www.gnu.org/licenses/>.

package caldav

import (
	"code.vikunja.io/api/pkg/utils"
	"strconv"
	"time"
)

// Event holds a single caldav event
type Event struct {
	Summary     string
	Description string
	UID         string
	Alarms      []Alarm

	TimestampUnix int64
	StartUnix     int64
	EndUnix       int64
}

// Alarm holds infos about an alarm from a caldav event
type Alarm struct {
	TimeUnix    int64
	Description string
}

// Config is the caldav calendar config
type Config struct {
	Name   string
	ProdID string
}

// ParseEvents parses an array of caldav events and gives them back as string
func ParseEvents(config *Config, events []*Event) (caldavevents string) {
	caldavevents += `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:` + config.Name + `
PRODID:-//` + config.ProdID + `//EN`

	for _, e := range events {

		if e.UID == "" {
			e.UID = makeCalDavTimeFromUnixTime(e.TimestampUnix) + utils.Sha256(e.Summary)
		}

		caldavevents += `
BEGIN:VEVENT
UID:` + e.UID + `
SUMMARY:` + e.Summary + `
DESCRIPTION:` + e.Description + `
DTSTAMP:` + makeCalDavTimeFromUnixTime(e.TimestampUnix) + `
DTSTART:` + makeCalDavTimeFromUnixTime(e.StartUnix) + `
DTEND:` + makeCalDavTimeFromUnixTime(e.EndUnix)

		for _, a := range e.Alarms {
			if a.Description == "" {
				a.Description = e.Summary
			}

			caldavevents += `
BEGIN:VALARM
TRIGGER:` + calcAlarmDateFromReminder(e.StartUnix, a.TimeUnix) + `
ACTION:DISPLAY
DESCRIPTION:` + a.Description + `
END:VALARM`
		}
		caldavevents += `
END:VEVENT`
	}

	caldavevents += `
END:VCALENDAR` // Need a line break

	return
}

func makeCalDavTimeFromUnixTime(unixtime int64) (caldavtime string) {
	tz, _ := time.LoadLocation("UTC")
	tm := time.Unix(unixtime, 0).In(tz)
	return tm.Format("20060102T150405")
}

func calcAlarmDateFromReminder(eventStartUnix, reminderUnix int64) (alarmTime string) {
	if eventStartUnix > reminderUnix {
		alarmTime += `-`
	}
	alarmTime += `PT`
	diff := eventStartUnix - reminderUnix
	if diff < 0 { // Make it positive
		diff = diff * -1
	}
	alarmTime += strconv.Itoa(int(diff/60)) + "M"
	return
}