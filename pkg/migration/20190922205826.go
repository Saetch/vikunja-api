// Copyright 2018-2020 Vikunja and contriubtors. All rights reserved.
//
// This file is part of Vikunja.
//
// Vikunja is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Vikunja is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Vikunja.  If not, see <https://www.gnu.org/licenses/>.

package migration

import (
	"code.vikunja.io/api/pkg/models"
	"github.com/go-xorm/xorm"
	"src.techknowlogick.com/xormigrate"
)

// TaskRelation represents a kind of relation between two tasks
type taskRelation20190922205826 struct {
	ID           int64               `xorm:"int(11) autoincr not null unique pk"`
	TaskID       int64               `xorm:"int(11) not null"`
	OtherTaskID  int64               `xorm:"int(11) not null"`
	RelationKind models.RelationKind `xorm:"varchar(50) not null"`
	CreatedByID  int64               `xorm:"int(11) not null"`
	Created      int64               `xorm:"created not null"`
}

// TableName holds the table name for the task relation table
func (taskRelation20190922205826) TableName() string {
	return "task_relations"
}

type task20190922205826 struct {
	ID           int64 `xorm:"int(11) autoincr not null unique pk"`
	CreatedByID  int64 `xorm:"int(11) not null"`
	ParentTaskID int64 `xorm:"int(11) INDEX null"`
}

func (task20190922205826) TableName() string {
	return "tasks"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20190922205826",
		Description: "Add task relations",
		Migrate: func(tx *xorm.Engine) error {
			err := tx.Sync2(taskRelation20190922205826{})
			if err != nil {
				return err
			}

			// Get all current subtasks and put them in a new table
			tasks := []*task20190922205826{}
			err = tx.Where("parent_task_id != null OR parent_task_id != 0").Find(&tasks)
			if err != nil {
				return err
			}

			var migratedRelations = make([]*taskRelation20190922205826, 0, len(tasks)*2)
			for _, t := range tasks {
				migratedRelations = append(migratedRelations,
					&taskRelation20190922205826{
						TaskID:       t.ID,
						OtherTaskID:  t.ParentTaskID,
						RelationKind: models.RelationKindParenttask,
						CreatedByID:  t.CreatedByID,
					},
					&taskRelation20190922205826{
						TaskID:       t.ParentTaskID,
						OtherTaskID:  t.ID,
						RelationKind: models.RelationKindSubtask,
						CreatedByID:  t.CreatedByID,
					})
			}

			_, err = tx.Insert(migratedRelations)
			if err != nil {
				return err
			}

			return dropTableColum(tx, "tasks", "parent_task_id")
		},
		Rollback: func(tx *xorm.Engine) error {
			return tx.DropTables(taskRelation20190922205826{})
		},
	})
}