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
import ReactTags from "react-tag-autocomplete";
import {randomClassName} from "../../utils";
import "./common.scss";
import "./TagField.scss";

const classNames = require("classnames");
const _ = require("lodash");

class TagField extends Component {

    state = {};

    constructor(props) {
        super(props);

        this.id = `field-${this.props.name || "noname"}-${randomClassName()}`;
    }

    onAddition = (tag) => {
        if (typeof this.props.onChange === "function") {
            this.props.onChange([].concat(this.props.tags, tag), true, tag); // true == addition
        }
    };

    onDelete = (i) => {
        if (typeof this.props.onChange === "function") {
            const tags = _.clone(this.props.tags);
            const tag = tags[i];
            tags.splice(i, 1);
            this.props.onChange(tags, true, tag);  // false == delete
        }
    };


    render() {
        const small = this.props.small || false;
        const name = this.props.name || null;

        return (
            <div className={classNames("field", "tag-field", {"field-small": small},
                {"field-inline": this.props.inline})}>
                {name &&
                <div className="field-name">
                    <label>{name}:</label>
                </div>
                }
                <div className="field-input">
                    <ReactTags {...this.props} tags={this.props.tags || []} autoresize={false}
                               onDelete={this.onDelete} onAddition={this.onAddition}
                               placeholderText={this.props.placeholder || ""}/>
                </div>
            </div>
        );
    }
}

export default TagField;
