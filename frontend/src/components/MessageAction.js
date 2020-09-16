import React, {Component} from 'react';
import './MessageAction.scss';
import {Button, FormControl, InputGroup, Modal} from "react-bootstrap";

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
        document.execCommand('copy');
        this.setState({copyButtonText: "copied!"});
        setTimeout(() => this.setState({copyButtonText: "copy"}), 3000);
    }

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
                    <InputGroup>
                        <FormControl as="textarea" className="message-action-value" readOnly={true}
                                     style={{"height": "300px"}} value={this.props.actionValue} ref={this.actionValue} />
                    </InputGroup>
                </Modal.Body>
                <Modal.Footer className="dialog-footer">
                    <Button variant="green" onClick={this.copyActionValue}>{this.state.copyButtonText}</Button>
                    <Button variant="red" onClick={this.props.onHide}>close</Button>
                </Modal.Footer>
            </Modal>
        );
    }
}

export default MessageAction;
