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

class CopyDialog extends Component {

    state = {
        copyButtonText: "copy"
    };

    constructor(props) {
        super(props);
        this.textbox = React.createRef();
    }

    copyActionValue = () => {
        this.textbox.current.select();
        document.execCommand("copy");
        this.setState({copyButtonText: "copied!"});
        this.timeoutHandle = setTimeout(() => this.setState({copyButtonText: "copy"}), 3000);
    };

    componentWillUnmount() {
        if (this.timeoutHandle) {
            clearTimeout(this.timeoutHandle);
        }
    }

    render() {
        return (
            <Modal show={true} size="lg" aria-labelledby="message-action-dialog" centered>
                <Modal.Header>
                    <Modal.Title id="message-action-dialog">
                        {this.props.name}
                    </Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <TextField readonly={this.props.readonly} value={this.props.value} textRef={this.textbox}
                               rows={15}/>
                </Modal.Body>
                <Modal.Footer className="dialog-footer">
                    <ButtonField variant="red" bordered onClick={this.props.onHide} name="close"/>
                    <ButtonField variant="green" bordered onClick={this.copyActionValue}
                                 name={this.state.copyButtonText}/>
                </Modal.Footer>
            </Modal>
        );
    }
}

export default CopyDialog;
