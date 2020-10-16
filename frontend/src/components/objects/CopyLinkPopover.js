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
import TextField from "../fields/TextField";
import LinkPopover from "./LinkPopover";

class CopyLinkPopover extends Component {

    state = {};

    constructor(props) {
        super(props);

        this.copyTextarea = React.createRef();
    }

    handleClick = () => {
        this.copyTextarea.current.select();
        document.execCommand("copy");
        this.setState({copiedMessage: true});
        setTimeout(() => this.setState({copiedMessage: false}), 3000);
    };

    render() {
        const copyPopoverContent = <div style={{"width": "250px"}}>
            {this.state.copiedMessage ? <span><strong>Copied!</strong></span> :
                <span>Click to <strong>copy</strong></span>}
            <TextField readonly rows={2} value={this.props.value} textRef={this.copyTextarea}/>
        </div>;

        return (
            <LinkPopover text={<span className={this.props.textClassName}
                                     onClick={this.handleClick}>{this.props.text}</span>}
                         content={copyPopoverContent} placement="right"/>
        );
    }
}

export default CopyLinkPopover;
