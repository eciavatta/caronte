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

const _ = require("lodash");

class Dispatcher {

    constructor() {
        this.listeners = [];
    }

    dispatch = (topic, payload) => {
        this.listeners.filter((l) => l.topic === topic).forEach((l) => l.callback(payload));
    };

    register = (topic, callback) => {
        if (typeof callback !== "function") {
            throw new Error("dispatcher callback must be a function");
        }
        if (typeof topic === "string") {
            this.listeners.push({topic, callback});
        } else if (typeof topic === "object" && Array.isArray(topic)) {
            topic.forEach((e) => {
                if (typeof e !== "string") {
                    throw new Error("all topics must be strings");
                }
            });

            topic.forEach((e) => this.listeners.push({e, callback}));
        } else {
            throw new Error("topic must be a string or an array of strings");
        }
    };

    unregister = (callback) => {
        _.remove(this.listeners, (l) => l.callback === callback);
    };

}

const dispatcher = new Dispatcher();

export default dispatcher;
