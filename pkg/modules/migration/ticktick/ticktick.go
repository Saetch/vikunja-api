// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package ticktick

import (
	"encoding/csv"
	"errors"
	"io"
	"sort"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/migration"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"

	"github.com/gocarina/gocsv"
)

const timeISO = "2006-01-02T15:04:05-0700"

type Migrator struct {
}

type tickTickTask struct {
	FolderName        string        `csv:"Folder Name"`
	ProjectName       string        `csv:"List Name"`
	Title             string        `csv:"Title"`
	TagsList          string        `csv:"Tags"`
	Tags              []string      `csv:"-"`
	Content           string        `csv:"Content"`
	IsChecklistString string        `csv:"Is Check list"`
	IsChecklist       bool          `csv:"-"`
	StartDate         tickTickTime  `csv:"Start Date"`
	DueDate           tickTickTime  `csv:"Due Date"`
	ReminderDuration  string        `csv:"Reminder"`
	Reminder          time.Duration `csv:"-"`
	Repeat            string        `csv:"Repeat"`
	Priority          int           `csv:"Priority"`
	Status            string        `csv:"Status"`
	CreatedTime       tickTickTime  `csv:"Created Time"`
	CompletedTime     tickTickTime  `csv:"Completed Time"`
	Order             float64       `csv:"Order"`
	TaskID            int64         `csv:"taskId"`
	ParentID          int64         `csv:"parentId"`
}

type tickTickTime struct {
	time.Time
}

func (date *tickTickTime) UnmarshalCSV(csv string) (err error) {
	date.Time = time.Time{}
	if csv == "" {
		return nil
	}
	date.Time, err = time.Parse(timeISO, csv)
	return err
}

func convertTickTickToVikunja(tasks []*tickTickTask) (result []*models.ProjectWithTasksAndBuckets) {
	parent := &models.ProjectWithTasksAndBuckets{
		Project: models.Project{
			Title: "Migrated from TickTick",
		},
		ChildProjects: []*models.ProjectWithTasksAndBuckets{},
	}

	projects := make(map[string]*models.ProjectWithTasksAndBuckets)
	for _, t := range tasks {
		_, has := projects[t.ProjectName]
		if !has {
			projects[t.ProjectName] = &models.ProjectWithTasksAndBuckets{
				Project: models.Project{
					Title: t.ProjectName,
				},
			}
		}

		labels := make([]*models.Label, 0, len(t.Tags))
		for _, tag := range t.Tags {
			labels = append(labels, &models.Label{
				Title: tag,
			})
		}

		task := &models.TaskWithComments{
			Task: models.Task{
				ID:          t.TaskID,
				Title:       t.Title,
				Description: t.Content,
				StartDate:   t.StartDate.Time,
				EndDate:     t.DueDate.Time,
				DueDate:     t.DueDate.Time,
				Done:        t.Status == "1",
				DoneAt:      t.CompletedTime.Time,
				Position:    t.Order,
				Labels:      labels,
			},
		}

		if !t.DueDate.IsZero() && t.Reminder > 0 {
			task.Task.Reminders = []*models.TaskReminder{
				{
					RelativeTo:     models.ReminderRelationDueDate,
					RelativePeriod: int64((t.Reminder * -1).Seconds()),
				},
			}
		}

		if t.ParentID != 0 {
			task.RelatedTasks = map[models.RelationKind][]*models.Task{
				models.RelationKindParenttask: {{ID: t.ParentID}},
			}
		}

		projects[t.ProjectName].Tasks = append(projects[t.ProjectName].Tasks, task)
	}

	for _, l := range projects {
		parent.ChildProjects = append(parent.ChildProjects, l)
	}

	sort.Slice(parent.ChildProjects, func(i, j int) bool {
		return parent.ChildProjects[i].Title < parent.ChildProjects[j].Title
	})

	return []*models.ProjectWithTasksAndBuckets{parent}
}

// Name is used to get the name of the ticktick migration - we're using the docs here to annotate the status route.
// @Summary Get migration status
// @Description Returns if the current user already did the migation or not. This is useful to show a confirmation message in the frontend if the user is trying to do the same migration again.
// @tags migration
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} migration.Status "The migration status"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /migration/ticktick/status [get]
func (m *Migrator) Name() string {
	return "ticktick"
}

func newLineSkipDecoder(r io.Reader, linesToSkip int) gocsv.SimpleDecoder {
	reader := csv.NewReader(r)
	//	reader.FieldsPerRecord = -1
	for i := 0; i < linesToSkip; i++ {
		_, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Debugf("[TickTick Migration] CSV parse error: %s", err)
		}
	}
	reader.FieldsPerRecord = 0
	return gocsv.NewSimpleDecoderFromCSVReader(reader)
}

// Migrate takes a ticktick export, parses it and imports everything in it into Vikunja.
// @Summary Import all projects, tasks etc. from a TickTick backup export
// @Description Imports all projects, tasks, notes, reminders, subtasks and files from a TickTick backup export into Vikunja.
// @tags migration
// @Accept x-www-form-urlencoded
// @Produce json
// @Security JWTKeyAuth
// @Param import formData string true "The TickTick backup csv file."
// @Success 200 {object} models.Message "A message telling you everything was migrated successfully."
// @Failure 500 {object} models.Message "Internal server error"
// @Router /migration/ticktick/migrate [post]
func (m *Migrator) Migrate(user *user.User, file io.ReaderAt, size int64) error {
	fr := io.NewSectionReader(file, 0, size)
	//r := csv.NewReader(fr)

	allTasks := []*tickTickTask{}
	decode := newLineSkipDecoder(fr, 3)
	err := gocsv.UnmarshalDecoder(decode, &allTasks)
	if err != nil {
		return err
	}

	for _, task := range allTasks {
		if task.IsChecklistString == "Y" {
			task.IsChecklist = true
		}

		reminder := utils.ParseISO8601Duration(task.ReminderDuration)
		if reminder > 0 {
			task.Reminder = reminder
		}

		task.Tags = strings.Split(task.TagsList, ", ")
	}

	vikunjaTasks := convertTickTickToVikunja(allTasks)

	return migration.InsertFromStructure(vikunjaTasks, user)
}
