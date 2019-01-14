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

package models

import (
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/web"
	"github.com/go-xorm/builder"
)

// CanUpdate checks if a user can update a label
func (l *Label) CanUpdate(a web.Auth) bool {
	return l.isLabelOwner(a) // Only owners should be allowed to update a label
}

// CanDelete checks if a user can delete a label
func (l *Label) CanDelete(a web.Auth) bool {
	return l.isLabelOwner(a) // Only owners should be allowed to delete a label
}

// CanRead checks if a user can read a label
func (l *Label) CanRead(a web.Auth) bool {
	return l.hasAccessToLabel(a)
}

// CanCreate checks if the user can create a label
// Currently a dummy.
func (l *Label) CanCreate(a web.Auth) bool {
	return true
}

func (l *Label) isLabelOwner(a web.Auth) bool {
	u := getUserForRights(a)
	lorig, err := getLabelByIDSimple(l.ID)
	if err != nil {
		log.Log.Errorf("Error occurred during isLabelOwner for Label: %v", err)
		return false
	}
	return lorig.CreatedByID == u.ID
}

// Helper method to check if a user can see a specific label
func (l *Label) hasAccessToLabel(a web.Auth) bool {
	u := getUserForRights(a)

	// Get all tasks
	taskIDs, err := getUserTaskIDs(u)
	if err != nil {
		log.Log.Errorf("Error occurred during hasAccessToLabel for Label: %v", err)
		return false
	}

	// Get all labels associated with these tasks
	var labels []*Label
	has, err := x.Table("labels").
		Select("labels.*").
		Join("LEFT", "label_task", "label_task.label_id = labels.id").
		Where("label_task.label_id != null OR labels.created_by_id = ?", u.ID).
		Or(builder.In("label_task.task_id", taskIDs)).
		And("labels.id = ?", l.ID).
		GroupBy("labels.id").
		Exist(&labels)
	if err != nil {
		log.Log.Errorf("Error occurred during hasAccessToLabel for Label: %v", err)
		return false
	}

	return has
}