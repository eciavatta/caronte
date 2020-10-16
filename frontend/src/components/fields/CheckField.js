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

import React, {Component} from "react";
import {randomClassName} from "../../utils";
import "./CheckField.scss";
import "./common.scss";

const classNames = require("classnames");

class CheckField extends Component {

    constructor(props) {
        super(props);

        this.id = `field-${this.props.name || "noname"}-${randomClassName()}`;
    }

    render() {
        const checked = this.props.checked || false;
        const small = this.props.small || false;
        const name = this.props.name || null;
        const handler = () => {
            if (!this.props.readonly && this.props.onChange) {
                this.props.onChange(!checked);
            }
        };

        return (
            <div className={classNames("field", "check-field", {"field-checked": checked}, {"field-small": small})}>
                <div className="field-input">
                    <input type="checkbox" id={this.id} checked={checked} onChange={handler}/>
                    <label htmlFor={this.id}>{(checked ? "✓ " : "✗ ") + (name != null ? name : "")}</label>
                </div>
            </div>
        );
    }
}

export default CheckField;
