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
import {Modal} from "react-bootstrap";
import ButtonField from "../fields/ButtonField";
import TextField from "../fields/TextField";
import "./MessageAction.scss";

class MessageAction extends Component {

    constructor(props) {
        super(props);
        this.state = {
            copyButtonText: "copy"
        };
        this.actionValue = React.createRef();
        this.copyActionValue = this.copyActionValue.bind(this);
    }

    copyActionValue() {
        this.actionValue.current.select();
        document.execCommand("copy");
        this.setState({copyButtonText: "copied!"});
        setTimeout(() => this.setState({copyButtonText: "copy"}), 3000);
    }

    render() {
        return (
            <Modal
                {...this.props}
                show={true}
                size="lg"
                aria-labelledby="message-action-dialog"
                centered
            >
                <Modal.Header>
                    <Modal.Title id="message-action-dialog">
                        {this.props.actionName}
                    </Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <TextField readonly value={this.props.actionValue} textRef={this.actionValue} rows={15}/>
                </Modal.Body>
                <Modal.Footer className="dialog-footer">
                    <ButtonField variant="green" bordered onClick={this.copyActionValue}
                                 name={this.state.copyButtonText}/>
                    <ButtonField variant="red" bordered onClick={this.props.onHide} name="close"/>
                </Modal.Footer>
            </Modal>
        );
    }
}

export default MessageAction;
