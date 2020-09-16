import React, {Component} from 'react';
import './MessageAction.scss';
import {Button, FormControl, InputGroup, Modal} from "react-bootstrap";

class MessageAction extends Component {



    render() {
        return (
            <Modal
                {...this.props}
                show="true"
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
                    <div className="message-action-value">
                        <pre>
                            {this.props.actionValue}
                        </pre>
                    </div>

                    {/*<InputGroup>*/}
                    {/*    <FormControl as="textarea" className="message-action-value" readOnly={true}*/}
                    {/*                 style={{"height": "300px"}}*/}
                    {/*                 value={this.props.actionValue}/>*/}
                    {/*</InputGroup>*/}
                </Modal.Body>
                <Modal.Footer className="dialog-footer">
                    <Button variant="green" onClick={this.copyActionValue}>copy</Button>
                    <Button variant="red" onClick={this.props.onHide}>close</Button>
                </Modal.Footer>
            </Modal>
        );
    }
}

export default MessageAction;
