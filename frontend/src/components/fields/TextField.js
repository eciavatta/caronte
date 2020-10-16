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
import "./common.scss";
import "./TextField.scss";

const classNames = require("classnames");

class TextField extends Component {

    constructor(props) {
        super(props);

        this.id = `field-${this.props.name || "noname"}-${randomClassName()}`;
    }

    render() {
        const name = this.props.name || null;
        const error = this.props.error || null;
        const rows = this.props.rows || 3;

        const handler = (e) => {
            if (this.props.onChange) {
                if (e == null) {
                    this.props.onChange("");
                } else {
                    this.props.onChange(e.target.value);
                }
            }
        };

        return (
            <div className={classNames("field", "text-field", {"field-active": this.props.active},
                {"field-invalid": this.props.invalid}, {"field-small": this.props.small})}>
                {name && <label htmlFor={this.id}>{name}:</label>}
                <textarea id={this.id} placeholder={this.props.defaultValue} onChange={handler} rows={rows}
                          readOnly={this.props.readonly} value={this.props.value} ref={this.props.textRef}/>
                {error && <div className="field-error">error: {error}</div>}
            </div>
        );
    }
}

export default TextField;
