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

import dispatcher from "./dispatcher";
import log from "./log";

class Notifications {

    constructor() {
        const location = document.location;
        this.wsUrl = `ws://${location.hostname}${location.port ? ":" + location.port : ""}/ws`;
    }

    createWebsocket = () => {
        this.ws = new WebSocket(this.wsUrl);
        this.ws.onopen = this.onWebsocketOpen;
        this.ws.onerror = this.onWebsocketError;
        this.ws.onclose = this.onWebsocketClose;
        this.ws.onmessage = this.onWebsocketMessage;
    };

    onWebsocketOpen = () => {
        log.debug("Connected to backend with websocket");
    };

    onWebsocketError = (err) => {
        this.ws.close();
        log.error("Websocket error", err);
        setTimeout(() => this.createWebsocket(), 3000);
    };

    onWebsocketClose = () => {
        log.debug("Closed websocket connection with backend");
    };

    onWebsocketMessage = (message) => {
        dispatcher.dispatch("notifications", JSON.parse(message.data));
    };
}

const notifications = new Notifications();

export default notifications;
