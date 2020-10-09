/*
 * This file is part of caronte (https://github.com/eciavatta/caronte).
 * Copyright (c) 2020 Emiliano Ciavatta.
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
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type NotificationController struct {
	upgrader           websocket.Upgrader
	clients            map[net.Addr]*client
	broadcast          chan interface{}
	register           chan *client
	unregister         chan *client
	applicationContext *ApplicationContext
}

func NewNotificationController(applicationContext *ApplicationContext) *NotificationController {
	return &NotificationController{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		clients:            make(map[net.Addr]*client),
		broadcast:          make(chan interface{}),
		register:           make(chan *client),
		unregister:         make(chan *client),
		applicationContext: applicationContext,
	}
}

type client struct {
	conn                   *websocket.Conn
	send                   chan interface{}
	notificationController *NotificationController
}

func (wc *NotificationController) NotificationHandler(w http.ResponseWriter, r *http.Request) error {
	conn, err := wc.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.WithError(err).Error("failed to set websocket upgrade")
		return err
	}

	client := &client{
		conn:                   conn,
		send:                   make(chan interface{}),
		notificationController: wc,
	}
	wc.register <- client
	go client.readPump()
	go client.writePump()

	return nil
}

func (wc *NotificationController) Run() {
	for {
		select {
		case client := <-wc.register:
			wc.clients[client.conn.RemoteAddr()] = client
			payload := gin.H{"event": "connected", "message": gin.H{
				"version":           wc.applicationContext.Version,
				"is_configured":     wc.applicationContext.IsConfigured,
				"connected_clients": len(wc.clients),
			}}
			client.send <- payload
			log.WithField("connected_clients", len(wc.clients)).
				WithField("remote_address", client.conn.RemoteAddr()).
				Info("[+] a websocket client connected")
		case client := <-wc.unregister:
			if _, ok := wc.clients[client.conn.RemoteAddr()]; ok {
				close(client.send)
				_ = client.conn.WriteMessage(websocket.CloseMessage, nil)
				_ = client.conn.Close()
				delete(wc.clients, client.conn.RemoteAddr())
				log.WithField("connected_clients", len(wc.clients)).
					WithField("remote_address", client.conn.RemoteAddr()).
					Info("[-] a websocket client disconnected")
			}
		case payload := <-wc.broadcast:
			for _, client := range wc.clients {
				select {
				case client.send <- payload:
				default:
					close(client.send)
					delete(wc.clients, client.conn.RemoteAddr())
				}
			}
		}
	}
}

func (wc *NotificationController) Notify(event string, message interface{}) {
	wc.broadcast <- gin.H{"event": event, "message": message}
}

func (c *client) readPump() {
	c.conn.SetReadLimit(maxMessageSize)
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		c.close()
		return
	}
	c.conn.SetPongHandler(func(string) error { return c.conn.SetReadDeadline(time.Now().Add(pongWait)) })
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.WithError(err).WithField("remote_address", c.conn.RemoteAddr()).
					Warn("unexpected websocket disconnection")
			}
			break
		}
	}

	c.close()
}

func (c *client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case payload, ok := <-c.send:
			if !ok {
				return
			}
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				c.close()
				return
			}
			if err := c.conn.WriteJSON(payload); err != nil {
				c.close()
				return
			}
		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				c.close()
				return
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.close()
				return
			}
		}
	}
}

func (c *client) close() {
	c.notificationController.unregister <- c
}
