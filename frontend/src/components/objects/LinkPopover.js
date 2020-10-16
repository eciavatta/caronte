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
import {randomClassName} from "../../utils";
import "./LinkPopover.scss";

class LinkPopover extends Component {

    constructor(props) {
        super(props);

        this.id = `link-overlay-${randomClassName()}`;
    }

    render() {
        const popover = (
            <Popover id={this.id}>
                {this.props.title && <Popover.Title as="h3">{this.props.title}</Popover.Title>}
                <Popover.Content>
                    {this.props.content}
                </Popover.Content>
            </Popover>
        );

        return (this.props.content ?
                <OverlayTrigger trigger={["hover", "focus"]} placement={this.props.placement || "top"}
                                overlay={popover}>
                    <span className="link-popover">{this.props.text}</span>
                </OverlayTrigger> :
                <span className="link-popover-empty">{this.props.text}</span>
        );
    }
}

export default LinkPopover;
