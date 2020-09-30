import React, {Component} from 'react';
import {randomClassName} from "../../utils";
import {OverlayTrigger, Popover} from "react-bootstrap";
import './LinkPopover.scss';

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
            <OverlayTrigger trigger={["hover", "focus"]} placement={this.props.placement || "top"} overlay={popover}>
                <span className="link-popover">{this.props.text}</span>
            </OverlayTrigger> :
            <span className="link-popover-empty">{this.props.text}</span>
        );
    }
}

export default LinkPopover;
