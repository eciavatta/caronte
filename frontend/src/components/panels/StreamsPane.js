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

import DOMPurify from "dompurify";
import React, {Component} from "react";
import {Row} from "react-bootstrap";
import ReactJson from "react-json-view";
import backend from "../../backend";
import log from "../../log";
import rules from "../../model/rules";
import {downloadBlob, getHeaderValue} from "../../utils";
import ButtonField from "../fields/ButtonField";
import ChoiceField from "../fields/ChoiceField";
import CopyDialog from "../dialogs/CopyDialog";
import "./StreamsPane.scss";

const reactStringReplace = require("react-string-replace");
const classNames = require("classnames");

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
        this.setState({messages: [], currentId: connectionId});
        backend.get(`/api/streams/${connectionId}?format=${this.state.format}`)
            .then((res) => this.setState({messages: res.json}));
    };

    setFormat = (format) => {
        if (this.validFormats.includes(format)) {
            this.setState({format});
        }
    };

    viewAs = (mode) => {
        if (mode === "decoded") {
            this.setState({tryParse: true});
        } else if (mode === "raw") {
            this.setState({tryParse: false});
        }
    };

    tryParseConnectionMessage = (connectionMessage) => {
        const isClient = connectionMessage["from_client"];
        if (connectionMessage.metadata == null) {
            return this.highlightRules(connectionMessage.content, isClient);
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
                    <div style={{"margin": "20px 0"}}>{this.highlightRules(m.body, isClient)}</div>
                    {unrollMap(m.trailers)}
                </span>;
            case "http-response":
                const contentType = getHeaderValue(m, "Content-Type");
                let body = m.body;
                if (contentType && contentType.includes("application/json")) {
                    try {
                        const json = JSON.parse(m.body);
                        if (typeof json === "object") {
                            body = <ReactJson src={json} theme="grayscale" collapsed={false} displayDataTypes={false}/>;
                        }
                    } catch (e) {
                        log.error(e);
                    }
                }

                return <span className="type-http-response">
                    <p style={{"marginBottom": "7px"}}>{m.protocol} <strong>{m.status}</strong></p>
                    {unrollMap(m.headers)}
                    <div style={{"margin": "20px 0"}}>{this.highlightRules(body, isClient)}</div>
                    {unrollMap(m.trailers)}
                </span>;
            default:
                return this.highlightRules(connectionMessage.content, isClient);
        }
    };

    highlightRules = (content, isClient) => {
        let streamContent = content;
        this.props.connection["matched_rules"].forEach(ruleId => {
            const rule = rules.ruleById(ruleId);
            rule.patterns.forEach(pattern => {
                if ((!isClient && pattern.direction === 1) || (isClient && pattern.direction === 2)) {
                    return;
                }
                let flags = "";
                pattern["caseless"] && (flags += "i");
                pattern["dot_all"] && (flags += "s");
                pattern["multi_line"] && (flags += "m");
                pattern["unicode_property"] && (flags += "u");
                const regex = new RegExp(pattern.regex.replace(/^\//, '(').replace(/\/$/, ')'), flags);
                streamContent = reactStringReplace(streamContent, regex, (match, i) => (
                    <span key={i} className="matched-occurrence" style={{"backgroundColor": rule.color}}>{match}</span>
                ));
            });
        });

        return streamContent;
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
                return Object.entries(connectionMessage.metadata["reproducers"]).map(([name, value]) =>
                    <ButtonField small key={name + "_button"} name={name} onClick={() => {
                        this.setState({
                            messageActionDialog: <CopyDialog actionName={name} value={value}
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
                .then((res) => downloadBlob(res.blob, `${this.state.currentId}-${value}-${this.state.format}.txt`))
                .catch((_) => log.error("Failed to download stream messages"));
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

        let payload = content.filter((c) => !this.state.tryParse || (this.state.tryParse && !c["is_metadata_continuation"]))
            .map((c, i) =>
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
            <div className="pane-container stream-pane">
                <div className="stream-pane-header container-fluid">
                    <Row>
                        <div className="header-info col">
                            <span><strong>flow</strong>: {conn["ip_src"]}:{conn["port_src"]} -> {conn["ip_dst"]}:{conn["port_dst"]}</span>
                            <span> | <strong>timestamp</strong>: {conn["started_at"]}</span>
                        </div>
                        <div className="header-actions col-auto">
                            <ChoiceField name="format" inline small onlyName
                                         keys={["default", "hex", "hexdump", "base32", "base64", "ascii", "binary", "decimal", "octal"]}
                                         values={["plain", "hex", "hexdump", "base32", "base64", "ascii", "binary", "decimal", "octal"]}
                                         onChange={this.setFormat} />

                            <ChoiceField name="view_as" inline small onlyName onChange={this.viewAs}
                                         keys={["decoded", "raw"]} values={["decoded", "raw"]} />

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
