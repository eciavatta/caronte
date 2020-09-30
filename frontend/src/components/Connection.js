import React, {Component} from 'react';
import './Connection.scss';
import {Form, OverlayTrigger, Popover} from "react-bootstrap";
import backend from "../backend";
import {dateTimeToTime, durationBetween, formatSize} from "../utils";
import ButtonField from "./fields/ButtonField";

const classNames = require('classnames');

class Connection extends Component {

    constructor(props) {
        super(props);
        this.state = {
            update: false,
            copiedMessage: false
        };

        this.copyTextarea = React.createRef();
        this.handleAction = this.handleAction.bind(this);
    }

    handleAction(name) {
        if (name === "hide") {
            const enabled = !this.props.data.hidden;
            backend.post(`/api/connections/${this.props.data.id}/${enabled ? "hide" : "show"}`)
                .then(_ => {
                    this.props.onEnabled(!enabled);
                    this.setState({update: true});
                });
        }
        if (name === "mark") {
            const marked = this.props.data.marked;
            backend.post(`/api/connections/${this.props.data.id}/${marked ? "unmark" : "mark"}`)
                .then(_ => {
                    this.props.onMarked(!marked);
                    this.setState({update: true});
                });
        }
        if (name === "copy") {
            this.copyTextarea.current.select();
            document.execCommand('copy');
            this.setState({copiedMessage: true});
            setTimeout(() => this.setState({copiedMessage: false}), 3000);
        }
    }

    render() {
        let conn = this.props.data;
        let serviceName = "/dev/null";
        let serviceColor = "#0F192E";
        if (conn.service.port !== 0) {
            serviceName = conn.service.name;
            serviceColor = conn.service.color;
        }
        let startedAt = new Date(conn.started_at);
        let closedAt = new Date(conn.closed_at);
        let processedAt = new Date(conn.processed_at);
        let timeInfo = <div>
            <span>Started at {startedAt.toLocaleDateString() + " " + startedAt.toLocaleTimeString()}</span><br/>
            <span>Processed at {processedAt.toLocaleDateString() + " " + processedAt.toLocaleTimeString()}</span><br/>
            <span>Closed at {closedAt.toLocaleDateString() + " " + closedAt.toLocaleTimeString()}</span>
        </div>;

        const popoverFor = function (name, content) {
            return <Popover id={`popover-${name}-${conn.id}`} className="connection-popover">
                <Popover.Content>
                    {content}
                </Popover.Content>
            </Popover>;
        };

        const commentPopoverContent = <div>
            <span>Click to <strong>{conn.comment.length > 0 ? "edit" : "add"}</strong> comment</span>
            {conn.comment && <Form.Control as="textarea" readOnly={true} rows={2} defaultValue={conn.comment}/>}
        </div>;

        const copyPopoverContent = <div>
            {this.state.copiedMessage ? <span><strong>Copied!</strong></span> :
                <span>Click to <strong>copy</strong> the connection id</span>}
            <Form.Control as="textarea" readOnly={true} rows={1} defaultValue={conn.id} ref={this.copyTextarea}/>
        </div>;

        return (
            <tr className={classNames("connection", {"connection-selected": this.props.selected},
                {"has-matched-rules": conn.matched_rules.length > 0})}>
                <td>
                    <span className="connection-service">
                        <ButtonField small fullSpan color={serviceColor} name={serviceName}
                                     onClick={() => this.props.addServicePortFilter(conn.port_dst)} />
                    </span>
                </td>
                <td className="clickable" onClick={this.props.onSelected}>{conn.ip_src}</td>
                <td className="clickable" onClick={this.props.onSelected}>{conn.port_src}</td>
                <td className="clickable" onClick={this.props.onSelected}>{conn.ip_dst}</td>
                <td className="clickable" onClick={this.props.onSelected}>{conn.port_dst}</td>
                <td className="clickable" onClick={this.props.onSelected}>{dateTimeToTime(conn.started_at)}</td>
                <td className="clickable" onClick={this.props.onSelected}>
                    <OverlayTrigger trigger={["focus", "hover"]} placement="right"
                                    overlay={popoverFor("duration", timeInfo)}>
                        <span className="test-tooltip">{durationBetween(startedAt, closedAt)}</span>
                    </OverlayTrigger>
                </td>
                <td className="clickable" onClick={this.props.onSelected}>{formatSize(conn.client_bytes)}</td>
                <td className="clickable" onClick={this.props.onSelected}>{formatSize(conn.server_bytes)}</td>
                <td>
                    {/*<OverlayTrigger trigger={["focus", "hover"]} placement="right"*/}
                    {/*                overlay={popoverFor("hide", <span>Hide this connection from the list</span>)}>*/}
                    {/*    <span className={"connection-icon" + (conn.hidden ? " icon-enabled" : "")}*/}
                    {/*          onClick={() => this.handleAction("hide")}>%</span>*/}
                    {/*</OverlayTrigger>*/}
                    <OverlayTrigger trigger={["focus", "hover"]} placement="right"
                                    overlay={popoverFor("hide", <span>Mark this connection</span>)}>
                        <span className={"connection-icon" + (conn.marked ? " icon-enabled" : "")}
                              onClick={() => this.handleAction("mark")}>!!</span>
                    </OverlayTrigger>
                    <OverlayTrigger trigger={["focus", "hover"]} placement="right"
                                    overlay={popoverFor("comment", commentPopoverContent)}>
                        <span className={"connection-icon" + (conn.comment ? " icon-enabled" : "")}
                              onClick={() => this.handleAction("comment")}>@</span>
                    </OverlayTrigger>
                    <OverlayTrigger trigger={["focus", "hover"]} placement="right"
                                    overlay={popoverFor("copy", copyPopoverContent)}>
                        <span className="connection-icon"
                              onClick={() => this.handleAction("copy")}>#</span>
                    </OverlayTrigger>
                </td>
            </tr>
        );
    }

}

export default Connection;
