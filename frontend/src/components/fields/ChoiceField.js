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
import "./ChoiceField.scss";
import "./common.scss";

const classNames = require("classnames");

class ChoiceField extends Component {

    constructor(props) {
        super(props);

        this.state = {
            expanded: false
        };

        this.id = `field-${this.props.name || "noname"}-${randomClassName()}`;
    }

    render() {
        const name = this.props.name || null;
        const inline = this.props.inline;

        const collapse = () => this.setState({expanded: false});
        const expand = () => this.setState({expanded: true});

        const handler = (key) => {
            collapse();
            if (this.props.onChange) {
                this.props.onChange(key);
            }
        };

        const keys = this.props.keys || [];
        const values = this.props.values || [];

        const options = keys.map((key, i) =>
            <span className="field-option" key={key} onClick={() => handler(key)}>{values[i]}</span>
        );

        let fieldValue = "";
        if (inline && name) {
            fieldValue = name;
        }
        if (!this.props.onlyName && inline && name) {
            fieldValue += ": ";
        }
        if (!this.props.onlyName) {
            fieldValue += this.props.value || "select a value";
        }

        return (
            <div className={classNames("field", "choice-field", {"field-inline": inline},
                {"field-small": this.props.small})}>
                {!inline && name && <label className="field-name">{name}:</label>}
                <div className={classNames("field-select", {"select-expanded": this.state.expanded})}
                     tabIndex={0} onBlur={collapse} onClick={() => this.state.expanded ? collapse() : expand()}>
                    <div className="field-value">{fieldValue}</div>
                    <div className="field-options">
                        {options}
                    </div>
                </div>
            </div>
        );
    }
}

export default ChoiceField;
