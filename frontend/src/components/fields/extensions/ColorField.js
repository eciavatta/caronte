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
import {OverlayTrigger, Popover} from "react-bootstrap";
import validation from "../../../validation";
import InputField from "../InputField";
import "./ColorField.scss";

class ColorField extends Component {

    constructor(props) {
        super(props);

        this.state = {};

        this.colors = ["#e53935", "#d81b60", "#8e24aa", "#5e35b1", "#3949ab", "#1e88e5", "#039be5", "#00acc1",
            "#00897b", "#43a047", "#7cb342", "#9e9d24", "#f9a825", "#fb8c00", "#f4511e", "#6d4c41"];
    }

    componentDidUpdate(prevProps, prevState, snapshot) {
        if (prevProps.value !== this.props.value) {
            this.onChange(this.props.value);
        }
    }

    onChange = (value) => {
        this.setState({invalid: value !== "" && !validation.isValidColor(value)});

        if (typeof this.props.onChange === "function") {
            this.props.onChange(value);
        }
    };

    render() {
        const colorButtons = this.colors.map((color) =>
            <span key={"button" + color} className="color-input" style={{"backgroundColor": color}}
                  onClick={() => {
                      if (typeof this.props.onChange === "function") {
                          this.props.onChange(color);
                      }
                      document.body.click(); // magic to close popup
                  }}/>);

        const popover = (
            <Popover id="popover-basic">
                <Popover.Title as="h3">choose a color</Popover.Title>
                <Popover.Content>
                    <div className="colors-container">
                        <div className="colors-row">
                            {colorButtons.slice(0, 8)}
                        </div>
                        <div className="colors-row">
                            {colorButtons.slice(8, 18)}
                        </div>
                    </div>
                </Popover.Content>
            </Popover>
        );

        let buttonStyles = {};
        if (this.props.value) {
            buttonStyles["backgroundColor"] = this.props.value;
        }

        return (
            <div className="field color-field">
                <div className="color-input">
                    <InputField {...this.props} onChange={this.onChange} invalid={this.state.invalid} name="color"
                                error={null}/>
                    <div className="color-picker">
                        <OverlayTrigger trigger="click" placement="top" overlay={popover} rootClose>
                            <button type="button" className="picker-button" style={buttonStyles}>pick</button>
                        </OverlayTrigger>
                    </div>
                </div>
                {this.props.error && <div className="color-error">{this.props.error}</div>}
            </div>
        );
    }

}

export default ColorField;
