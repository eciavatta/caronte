import React, {Component} from 'react';
import './PcapPane.scss';
import Table from "react-bootstrap/Table";
import backend from "../../backend";
import {createCurlCommand, formatSize, timestampToTime2} from "../../utils";
import {Button, Col, Container, Form, Row} from "react-bootstrap";
import InputField from "../fields/InputField";
import CheckField from "../fields/CheckField";
import TextField from "../fields/TextField";

class PcapPane extends Component {

    constructor(props) {
        super(props);

        this.state = {
            sessions: [],
            isFileValid: true,
            isFileFocused: false,
            selectedFile: null,
            uploadFlushAll: false,
            uploadStatusCode: null,
            uploadOutput: null
        };

        this.loadSessions = this.loadSessions.bind(this);
        this.handleFileChange = this.handleFileChange.bind(this);
        this.handleUploadPcap = this.handleUploadPcap.bind(this);
    }

    componentDidMount() {
        this.loadSessions();
    }

    loadSessions() {
        backend.get("/api/pcap/sessions").then(res => this.setState({sessions: res}));
    }

    handleFileChange(file) {
        this.setState({
            isFileValid: file != null && file.type.endsWith("pcap"),
            isFileFocused: false,
            selectedFile: file
        });
    }

    handleUploadPcap() {
        if (this.state.selectedFile == null || !this.state.isFileValid) {
            this.setState({isFileFocused: true});
            return;
        }

        const formData = new FormData();
        formData.append(
            "file",
            this.state.selectedFile
        );

        backend.postFile("/api/pcap/upload", formData).then(response =>
            response.json().then(result => this.setState({
                uploadStatusCode: response.status + " " + response.statusText,
                uploadOutput: JSON.stringify(result)
            }))
        );
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
                <td className="table-cell-action"><a target="_blank"
                                                     href={"/api/pcap/sessions/" + s["id"] + "/download"}>download</a>
                </td>
            </tr>
        );

        const uploadOutput = this.state.uploadOutput != null ? this.state.uploadOutput :
            createCurlCommand("pcap/upload", "POST", null, {
                file: "@" + ((this.state.selectedFile != null && this.state.isFileValid) ? this.state.selectedFile.name :
                    "invalid.pcap"),
                flush_all: this.state.uploadFlushAll
            })
        ;

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
                                    <span className="api-response">{this.state.uploadStatusCode}</span>
                                </div>

                                <div className="section-content">
                                    <InputField type={"file"} name={"file"} invalid={!this.state.isFileValid}
                                                active={this.state.isFileFocused}
                                                onChange={this.handleFileChange} value={this.state.selectedFile}
                                                defaultValue={"No .pcap[ng] selected"}/>

                                    <div className="upload-actions">
                                        <div className="upload-options">
                                            <span>options:</span>
                                            <CheckField name="flush_all" checked={this.state.uploadFlushAll}
                                                        onChange={v => this.setState({uploadFlushAll: v})}/>
                                        </div>
                                        <Button variant="green" onClick={this.handleUploadPcap}>upload</Button>
                                    </div>

                                    <TextField value={uploadOutput} rows={4} readonly small={true}/>
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
