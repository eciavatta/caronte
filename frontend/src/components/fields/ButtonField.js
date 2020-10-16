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
import "./ButtonField.scss";
import "./common.scss";

const classNames = require("classnames");

class ButtonField extends Component {

    render() {
        const handler = () => {
            if (typeof this.props.onClick === "function") {
                this.props.onClick();
            }
        };

        let buttonClassnames = {
            "button-bordered": this.props.bordered,
        };
        if (this.props.variant) {
            buttonClassnames[`button-variant-${this.props.variant}`] = true;
        }

        let buttonStyle = {};
        if (this.props.color) {
            buttonStyle["backgroundColor"] = this.props.color;
        }
        if (this.props.border) {
            buttonStyle["borderColor"] = this.props.border;
        }
        if (this.props.fullSpan) {
            buttonStyle["width"] = "100%";
        }
        if (this.props.rounded) {
            buttonStyle["borderRadius"] = "3px";
        }
        if (this.props.inline) {
            buttonStyle["marginTop"] = "8px";
        }

        return (
            <div className={classNames("field", "button-field", {"field-small": this.props.small},
                {"field-active": this.props.active})}>
                <button type="button" className={classNames(buttonClassnames)}
                        onClick={handler} style={buttonStyle} disabled={this.props.disabled}>{this.props.name}</button>
            </div>
        );
    }
}

export default ButtonField;
