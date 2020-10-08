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

import React, {Component} from 'react';
import './StreamsPane.scss';
import {Row} from 'react-bootstrap';
import MessageAction from "../objects/MessageAction";
import backend from "../../backend";
import ButtonField from "../fields/ButtonField";
import ChoiceField from "../fields/ChoiceField";
import DOMPurify from 'dompurify';
import ReactJson from 'react-json-view'
import {downloadBlob, getHeaderValue} from "../../utils";
import log from "../../log";

const classNames = require('classnames');

class StreamsPane extends Component {

    state = {
        messages: [],
        format: "default",
        tryParse: true
    };

    constructor(props) {
        super(props);

        this.validFormats = ["default", "hex", "hexdump", "base32", "base64", "ascii", "binary", "decimal", "octal"];
    }

    componentDidMount() {
        if (this.props.connection && this.state.currentId !== this.props.connection.id) {
            this.setState({currentId: this.props.connection.id});
            this.loadStream(this.props.connection.id);
        }

        document.title = "caronte:~/$";
    }

    componentDidUpdate(prevProps, prevState, snapshot) {
        if (this.props.connection && (
            this.props.connection !== prevProps.connection || this.state.format !== prevState.format)) {
            this.closeRenderWindow();
            this.loadStream(this.props.connection.id);
        }
    }

    componentWillUnmount() {
        this.closeRenderWindow();
    }

    loadStream = (connectionId) => {
        this.setState({messages: []});
        backend.get(`/api/streams/${connectionId}?format=${this.state.format}`)
            .then(res => this.setState({messages: res.json}));
    };

    setFormat = (format) => {
        if (this.validFormats.includes(format)) {
            this.setState({format: format});
        }
    };

    tryParseConnectionMessage = (connectionMessage) => {
        if (connectionMessage.metadata == null) {
            return connectionMessage.content;
        }
        if (connectionMessage["is_metadata_continuation"]) {
            return <span style={{"fontSize": "12px"}}>**already parsed in previous messages**</span>;
        }

        let unrollMap = (obj) => obj == null ? null : Object.entries(obj).map(([key, value]) =>
            <p key={key}><strong>{key}</strong>: {value}</p>
        );

        let m = connectionMessage.metadata;
        switch (m.type) {
            case "http-request":
                let url = <i><u><a href={"http://" + m.host + m.url} target="_blank"
                                   rel="noopener noreferrer">{m.host}{m.url}</a></u></i>;
                return <span className="type-http-request">
                    <p style={{"marginBottom": "7px"}}><strong>{m.method}</strong> {url} {m.protocol}</p>
                    {unrollMap(m.headers)}
                    <div style={{"margin": "20px 0"}}>{m.body}</div>
                    {unrollMap(m.trailers)}
                </span>;
            case "http-response":
                const contentType = getHeaderValue(m, "Content-Type");
                let body = m.body;
                if (contentType && contentType.includes("application/json")) {
                    try {
                        const json = JSON.parse(m.body);
                        body = <ReactJson src={json} theme="grayscale" collapsed={false} displayDataTypes={false}/>;
                    } catch (e) {
                        console.log(e);
                    }
                }

                return <span className="type-http-response">
                    <p style={{"marginBottom": "7px"}}>{m.protocol} <strong>{m.status}</strong></p>
                    {unrollMap(m.headers)}
                    <div style={{"margin": "20px 0"}}>{body}</div>
                    {unrollMap(m.trailers)}
                </span>;
            default:
                return connectionMessage.content;
        }
    };

    connectionsActions = (connectionMessage) => {
        if (!connectionMessage.metadata) {
            return null;
        }

        const m = connectionMessage.metadata;
        switch (m.type) {
            case "http-request" :
                if (!connectionMessage.metadata["reproducers"]) {
                    return;
                }
                return Object.entries(connectionMessage.metadata["reproducers"]).map(([actionName, actionValue]) =>
                    <ButtonField small key={actionName + "_button"} name={actionName} onClick={() => {
                        this.setState({
                            messageActionDialog: <MessageAction actionName={actionName} actionValue={actionValue}
                                                                onHide={() => this.setState({messageActionDialog: null})}/>
                        });
                    }}/>
                );
            case "http-response":
                const contentType = getHeaderValue(m, "Content-Type");

                if (contentType && contentType.includes("text/html")) {
                    return <ButtonField small name="render_html" onClick={() => {
                        let w;
                        if (this.state.renderWindow && !this.state.renderWindow.closed) {
                            w = this.state.renderWindow;
                        } else {
                            w = window.open("", "", "width=900, height=600, scrollbars=yes");
                            this.setState({renderWindow: w});
                        }
                        w.document.body.innerHTML = DOMPurify.sanitize(m.body);
                        w.focus();
                    }}/>;
                }
                break;
            default:
                return null;
        }
    };

    downloadStreamRaw = (value) => {
        if (this.state.currentId) {
            backend.download(`/api/streams/${this.props.connection.id}/download?format=${this.state.format}&type=${value}`)
                .then(res => downloadBlob(res.blob, `${this.state.currentId}-${value}-${this.state.format}.txt`))
                .catch(_ => log.error("Failed to download stream messages"));
        }
    };

    closeRenderWindow = () => {
        if (this.state.renderWindow) {
            this.state.renderWindow.close();
        }
    };

    render() {
        const conn = this.props.connection || {
            "ip_src": "0.0.0.0",
            "ip_dst": "0.0.0.0",
            "port_src": "0",
            "port_dst": "0",
            "started_at": new Date().toISOString(),
        };
        const content = this.state.messages || [];

        let payload = content.map((c, i) =>
            <div key={`content-${i}`}
                 className={classNames("connection-message", c["from_client"] ? "from-client" : "from-server")}>
                <div className="connection-message-header container-fluid">
                    <div className="row">
                        <div className="connection-message-info col">
                            <span><strong>offset</strong>: {c.index}</span> | <span><strong>timestamp</strong>: {c.timestamp}
                        </span> | <span><strong>retransmitted</strong>: {c["is_retransmitted"] ? "yes" : "no"}</span>
                        </div>
                        <div className="connection-message-actions col-auto">{this.connectionsActions(c)}</div>
                    </div>
                </div>
                <div className="connection-message-label">{c["from_client"] ? "client" : "server"}</div>
                <div
                    className="message-content">
                    {this.state.tryParse && this.state.format === "default" ? this.tryParseConnectionMessage(c) : c.content}
                </div>
            </div>
        );

        return (
            <div className="connection-content">
                <div className="connection-content-header container-fluid">
                    <Row>
                        <div className="header-info col">
                            <span><strong>flow</strong>: {conn["ip_src"]}:{conn["port_src"]} -> {conn["ip_dst"]}:{conn["port_dst"]}</span>
                            <span> | <strong>timestamp</strong>: {conn["started_at"]}</span>
                        </div>
                        <div className="header-actions col-auto">
                            <ChoiceField name="format" inline small onlyName
                                         keys={["default", "hex", "hexdump", "base32", "base64", "ascii", "binary", "decimal", "octal"]}
                                         values={["plain", "hex", "hexdump", "base32", "base64", "ascii", "binary", "decimal", "octal"]}
                                         onChange={this.setFormat} value={this.state.value}/>

                            <ChoiceField name="view_as" inline small onlyName keys={["default"]} values={["default"]}/>

                            <ChoiceField name="download_as" inline small onlyName onChange={this.downloadStreamRaw}
                                         keys={["nl_separated", "only_client", "only_server", "pwntools"]}
                                         values={["nl_separated", "only_client", "only_server", "pwntools"]}/>
                        </div>
                    </Row>
                </div>

                <pre>{payload}</pre>
                {this.state.messageActionDialog}
            </div>
        );
    }

}


export default StreamsPane;
