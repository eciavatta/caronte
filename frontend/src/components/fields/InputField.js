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
import "./InputField.scss";

const classNames = require("classnames");

class InputField extends Component {

    constructor(props) {
        super(props);

        this.id = `field-${this.props.name || "noname"}-${randomClassName()}`;
    }

    render() {
        const active = this.props.active || false;
        const invalid = this.props.invalid || false;
        const small = this.props.small || false;
        const inline = this.props.inline || false;
        const name = this.props.name || null;
        const value = this.props.value || "";
        const defaultValue = this.props.defaultValue || "";
        const type = this.props.type || "text";
        const error = this.props.error || null;

        const handler = (e) => {
            if (typeof this.props.onChange === "function") {
                if (type === "file") {
                    let file = e.target.files[0];
                    this.props.onChange(file);
                } else if (e == null) {
                    this.props.onChange(defaultValue);
                } else {
                    this.props.onChange(e.target.value);
                }
            }
        };
        let inputProps = {};
        if (type !== "file") {
            inputProps["value"] = value || defaultValue;
        }

        return (
            <div className={classNames("field", "input-field", {"field-active": active},
                {"field-invalid": invalid}, {"field-small": small}, {"field-inline": inline})}>
                <div className="field-wrapper">
                    {name &&
                    <div className="field-name">
                        <label>{name}:</label>
                    </div>
                    }
                    <div className="field-input">
                        <div className="field-value">
                            {type === "file" && <label for={this.id} className={"file-label"}>
                                {value.name || this.props.placeholder}</label>}
                            <input type={type} placeholder={this.props.placeholder} id={this.id}
                                   aria-describedby={this.id} onChange={handler} {...inputProps}
                                   readOnly={this.props.readonly}/>
                        </div>
                        {type !== "file" && value !== "" && !this.props.readonly &&
                        <div className="field-clear">
                            <span onClick={() => handler(null)}>del</span>
                        </div>
                        }
                    </div>
                </div>
                {error &&
                <div className="field-error">
                    error: {error}
                </div>
                }
            </div>
        );
    }
}

export default InputField;
