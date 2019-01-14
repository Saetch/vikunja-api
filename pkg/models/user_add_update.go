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
	"code.vikunja.io/api/pkg/mail"
	"code.vikunja.io/api/pkg/metrics"
	"code.vikunja.io/api/pkg/utils"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

// CreateUser creates a new user and inserts it into the database
func CreateUser(user User) (newUser User, err error) {

	newUser = user

	// Check if we have all needed informations
	if newUser.Password == "" || newUser.Username == "" {
		return User{}, ErrNoUsernamePassword{}
	}

	// Check if the user already existst with that username
	exists := true
	existingUser, err := GetUser(User{Username: newUser.Username})
	if err != nil {
		if IsErrUserDoesNotExist(err) {
			exists = false
		} else {
			return User{}, err
		}
	}
	if exists {
		return User{}, ErrUsernameExists{newUser.ID, newUser.Username}
	}

	// Check if the user already existst with that email
	exists = true
	existingUser, err = GetUser(User{Email: newUser.Email})
	if err != nil {
		if IsErrUserDoesNotExist(err) {
			exists = false
		} else {
			return User{}, err
		}
	}
	if exists {
		return User{}, ErrUserEmailExists{existingUser.ID, existingUser.Email}
	}

	// Hash the password
	newUser.Password, err = hashPassword(user.Password)
	if err != nil {
		return User{}, err
	}

	newUser.IsActive = true
	if viper.GetBool("mailer.enabled") {
		// The new user should not be activated until it confirms his mail address
		newUser.IsActive = false
		// Generate a confirm token
		newUser.EmailConfirmToken = utils.MakeRandomString(400)
	}

	// Insert it
	_, err = x.Insert(newUser)
	if err != nil {
		return User{}, err
	}

	// Update the metrics
	metrics.UpdateCount(1, metrics.ActiveUsersKey)

	// Get the  full new User
	newUserOut, err := GetUser(newUser)
	if err != nil {
		return User{}, err
	}

	// Create the user's namespace
	newN := &Namespace{Name: newUserOut.Username, Description: newUserOut.Username + "'s namespace.", Owner: newUserOut}
	err = newN.Create(&newUserOut)
	if err != nil {
		return User{}, err
	}

	// Dont send a mail if we're testing
	if !viper.GetBool("mailer.enabled") {
		return newUserOut, err
	}

	// Send the user a mail with a link to confirm the mail
	data := map[string]interface{}{
		"User": newUserOut,
	}

	mail.SendMailWithTemplate(user.Email, newUserOut.Username+" + Vikunja = <3", "confirm-email", data)

	return newUserOut, err
}

// HashPassword hashes a password
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// UpdateUser updates a user
func UpdateUser(user User) (updatedUser User, err error) {

	// Check if it exists
	theUser, err := GetUserByID(user.ID)
	if err != nil {
		return User{}, err
	}

	// Check if we have at least a username
	if user.Username == "" {
		//return User{}, ErrNoUsername{user.ID}
		user.Username = theUser.Username // Dont change the username if we dont have one
	}

	user.Password = theUser.Password // set the password to the one in the database to not accedently resetting it

	// Update it
	_, err = x.Id(user.ID).Update(user)
	if err != nil {
		return User{}, err
	}

	// Get the newly updated user
	updatedUser, err = GetUserByID(user.ID)
	if err != nil {
		return User{}, err
	}

	return updatedUser, err
}

// UpdateUserPassword updates the password of a user
func UpdateUserPassword(user *User, newPassword string) (err error) {

	// Get all user details
	theUser, err := GetUserByID(user.ID)
	if err != nil {
		return err
	}

	// Hash the new password and set it
	hashed, err := hashPassword(newPassword)
	if err != nil {
		return err
	}
	theUser.Password = hashed

	// Update it
	_, err = x.Id(user.ID).Update(theUser)
	if err != nil {
		return err
	}

	return err
}