import React, {Component} from 'react';
import './PcapPane.scss';
import Table from "react-bootstrap/Table";
import backend from "../../backend";
import {formatSize, timestampToTime2} from "../../utils";
import {Container, Row, Col, Form} from "react-bootstrap";
import StringField from "../fields/StringField";
import BooleanField from "../fields/BooleanField";

class PcapPane extends Component {

    constructor(props) {
        super(props);

        this.state = {
            sessions: [],
            test: false
        };

        this.loadSessions = this.loadSessions.bind(this);
    }

    componentDidMount() {
        this.loadSessions();
    }

    loadSessions() {
        backend.get("/api/pcap/sessions").then(res => this.setState({sessions: res}));
    }

    render() {
        let sessions = this.state.sessions.map(s =>
            <tr className="table-row">
                <td>{s["id"].substring(0, 8)}</td>
                <td>{timestampToTime2(s["started_at"])}</td>
                <td>{((new Date(s["completed_at"]) - new Date(s["started_at"])) / 1000).toFixed(3)}s</td>
                <td>{formatSize(s["size"])}</td>
                <td>{s["processed_packets"]}</td>
                <td>{s["invalid_packets"]}</td>
                <td>undefined</td>
                <td className="table-cell-action"><a target="_blank" href={"/api/pcap/sessions/" + s["id"] + "/download"}>download</a></td>
            </tr>
        );

        return (
            <div className="pane-container">
                <div className="pane-section">
                    <div className="section-header">
                        <span className="api-request">GET /api/pcap/sessions</span>
                        <span className="api-response">200 OK</span>
                    </div>

                    <div className="section-table">
                        <Table borderless size="sm">
                            <thead>
                            <tr>
                                <th>id</th>
                                <th>started_at</th>
                                <th>duration</th>
                                <th>size</th>
                                <th>processed_packets</th>
                                <th>invalid_packets</th>
                                <th>packets_per_service</th>
                                <th>actions</th>
                            </tr>
                            </thead>
                            <tbody>
                            {sessions}
                            </tbody>
                        </Table>
                    </div>
                </div>

                <div className="pane-section">
                    <Container className="p-0">
                        <Row>
                            <Col>
                                <div className="section-header">
                                    <span className="api-request">POST /api/pcap/upload</span>
                                    <span className="api-response"></span>
                                </div>

                                <div className="section-content">
                                    <Form.File className="custom-file" onChange={this.onFileChange}
                                        label=".pcap/.pcapng" id="custom-file"
                                        custom={true}
                                    />


                                    <br/><br/><br/><br/>
                                    <BooleanField small={true} name={"marked"} checked={this.state.test} onChange={(v) => this.setState({test: v})} />

                                </div>
                            </Col>

                            <Col>
                                <div className="section-header">
                                    <span className="api-request">POST /api/pcap/file</span>
                                    <span className="api-response"></span>
                                </div>

                                <div className="section-content">
                                    <Form.Control type="text" id="pcap-upload" className="custom-file"
                                                  onChange={this.onLocalFileChange} placeholder="local .pcap/.pcapng"
                                                  custom
                                    />
                                </div>
                            </Col>
                        </Row>
                    </Container>




                </div>

            </div>

        );
    }
}

export default PcapPane;
