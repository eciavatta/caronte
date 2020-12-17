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
import backend from "../../backend";
import log from "../../log";
import ButtonField from "../fields/ButtonField";
import TextField from "../fields/TextField";

class CommentDialog extends Component {

    state = {};

    componentDidMount() {
        this.setState({comment: this.props.initialComment || ""});
    }

    setComment = () => {
        if (this.state.comment === this.props.initialComment) {
            return this.close();
        }
        const comment = this.state.comment || null;
        backend.post(`/api/connections/${this.props.connectionId}/comment`, {comment})
            .then((_) => {
                this.close();
            }).catch((e) => {
                log.error(e);
                this.setState({error: "failed to save comment"});
            });
    };

    close = () => this.props.onSave(this.state.comment || null);

    render() {
        return (
            <Modal show size="md" aria-labelledby="comment-dialog" centered>
                <Modal.Header>
                    <Modal.Title id="comment-dialog">
                        ~/.comment
                    </Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <TextField value={this.state.comment} onChange={(comment) => this.setState({comment})}
                               rows={7} error={this.state.error}/>
                </Modal.Body>
                <Modal.Footer className="dialog-footer">
                    <ButtonField variant="red" bordered onClick={this.close} name="cancel"/>
                    <ButtonField variant="green" bordered onClick={this.setComment} name="save"/>
                </Modal.Footer>
            </Modal>
        );
    }
}

export default CommentDialog;
