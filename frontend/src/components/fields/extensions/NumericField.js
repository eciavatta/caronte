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
import InputField from "../InputField";

class NumericField extends Component {

    constructor(props) {
        super(props);

        this.state = {
            invalid: false
        };
    }

    componentDidUpdate(prevProps, prevState, snapshot) {
        if (prevProps.value !== this.props.value) {
            this.onChange(this.props.value);
        }
    }

    onChange = (value) => {
        value = value.toString().replace(/[^\d]/gi, "");
        let intValue = 0;
        if (value !== "") {
            intValue = parseInt(value, 10);
        }
        const valid =
            (!this.props.validate || (typeof this.props.validate === "function" && this.props.validate(intValue))) &&
            (!this.props.min || (typeof this.props.min === "number" && intValue >= this.props.min)) &&
            (!this.props.max || (typeof this.props.max === "number" && intValue <= this.props.max));
        this.setState({invalid: !valid});
        if (typeof this.props.onChange === "function") {
            this.props.onChange(intValue);
        }
    };

    render() {
        return (
            <InputField {...this.props} onChange={this.onChange} defaultValue={this.props.defaultValue || "0"}
                        invalid={this.state.invalid}/>
        );
    }

}

export default NumericField;
