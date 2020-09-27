import React, {Component} from 'react';
import './MessageAction.scss';
import {Modal} from "react-bootstrap";
import TextField from "./fields/TextField";
import ButtonField from "./fields/ButtonField";

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
                    <TextField readonly value={this.props.actionValue} textRef={this.actionValue} rows={15} />
                </Modal.Body>
                <Modal.Footer className="dialog-footer">
                    <ButtonField variant="green" onClick={this.copyActionValue} name={this.state.copyButtonText} />
                    <ButtonField variant="red" onClick={this.props.onHide} name="close" />
                </Modal.Footer>
            </Modal>
        );
    }
}

export default MessageAction;
