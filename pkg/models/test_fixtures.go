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
	"gopkg.in/testfixtures.v2"
)

var fixtures *testfixtures.Context

// InitFixtures initialize test fixtures for a test database
func InitFixtures(helper testfixtures.Helper, dir string) (err error) {
	testfixtures.SkipDatabaseNameCheck(true)
	fixtures, err = testfixtures.NewFolder(x.DB().DB, helper, dir)
	return err
}

// LoadFixtures load fixtures for a test database
func LoadFixtures() error {
	return fixtures.Load()
}