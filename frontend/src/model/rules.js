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

import backend from "../backend";
import log from "../log";

const _ = require("lodash");

class Rules {

    constructor() {
        this.rules = [];
        this.loadRules();
    }

    loadRules = () => {
        backend.get("/api/rules").then((res) => this.rules = res.json)
            .catch((err) => log.error("Failed to load rules", err));
    };

    allRules = () => _.clone(this.rules);

    ruleById = (id) => _.clone(this.rules.find(r => r.id === id));

}

const rules = new Rules();

export default rules;
