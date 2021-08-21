/*
 * This file is part of caronte (https://github.com/eciavatta/caronte).
 * Copyright (c) 2021 Emiliano Ciavatta.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, version 3.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type TestNotificationController struct {
	notificationChannel chan gin.H
}

func NewTestNotificationController() *TestNotificationController {
	return &TestNotificationController{
		notificationChannel: make(chan gin.H),
	}
}

func (wc *TestNotificationController) NotificationHandler(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (wc *TestNotificationController) Run() {
}

func (wc *TestNotificationController) Notify(event string, message interface{}) {
	wc.notificationChannel <- gin.H{"event": event, "message": message}
}
