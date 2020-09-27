import React, {Component} from 'react';
import './PcapPane.scss';
import Table from "react-bootstrap/Table";
import backend from "../../backend";
import {createCurlCommand, formatSize, timestampToTime2} from "../../utils";
import {Col, Container, Row} from "react-bootstrap";
import InputField from "../fields/InputField";
import CheckField from "../fields/CheckField";
import TextField from "../fields/TextField";
import ButtonField from "../fields/ButtonField";

class PcapPane extends Component {

    constructor(props) {
        super(props);

        this.state = {
            sessions: [],
            isUploadFileValid: true,
            isUploadFileFocused: false,
            uploadSelectedFile: null,
            uploadFlushAll: false,
            uploadStatusCode: null,
            uploadOutput: null,
            isFileValid: true,
            isFileFocused: false,
            fileValue: "",
            fileFlushAll: false,
            fileStatusCode: null,
            fileOutput: null,
            deleteOriginalFile: false
        };
    }

    componentDidMount() {
        this.loadSessions();
    }

    loadSessions = () => {
        backend.getJson("/api/pcap/sessions").then(res => this.setState({sessions: res}));
    };

    handleUploadFileChange = (file) => {
        this.setState({
            isUploadFileValid: file != null && file.type.endsWith("pcap"),
            isUploadFileFocused: false,
            uploadSelectedFile: file
        });
    };

    handleUploadPcap = () => {
        if (this.state.uploadSelectedFile == null || !this.state.isUploadFileValid) {
            this.setState({isUploadFileFocused: true});
            return;
        }

        const formData = new FormData();
        formData.append(
            "file",
            this.state.uploadSelectedFile
        );

        backend.postFile("/api/pcap/upload", formData).then(response =>
            response.json().then(result => this.setState({
                uploadStatusCode: response.status + " " + response.statusText,
                uploadOutput: JSON.stringify(result)
            }))
        );
    };

    handleFileChange = (file) => {
        this.setState({
            isFileValid: file !== "" && file.endsWith("pcap"),
            isFileFocused: false,
            fileValue: file
        });
    };

    handleProcessPcap = () => {
        if (this.state.fileValue === "" || !this.state.isFileValid) {
            this.setState({isFileFocused: true});
            return;
        }

        backend.post("/api/pcap/file", {
            file: this.state.fileValue,
            flush_all: this.state.fileFlushAll,
            delete_original_file: this.state.deleteOriginalFile
        }).then(response =>
            response.json().then(result => this.setState({
                fileStatusCode: response.status + " " + response.statusText,
                fileOutput: JSON.stringify(result)
            }))
        );
    };

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
                file: "@" + ((this.state.uploadSelectedFile != null && this.state.isUploadFileValid) ? this.state.uploadSelectedFile.name :
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
                            <Col style={{"paddingRight": "0"}}>
                                <div className="section-header">
                                    <span className="api-request">POST /api/pcap/upload</span>
                                    <span className="api-response">{this.state.uploadStatusCode}</span>
                                </div>

                                <div className="section-content">
                                    <InputField type={"file"} name={"file"} invalid={!this.state.isUploadFileValid}
                                                active={this.state.isUploadFileFocused}
                                                onChange={this.handleUploadFileChange} value={this.state.uploadSelectedFile}
                                                defaultValue={"no .pcap[ng] selected"}/>

                                    <div className="upload-actions">
                                        <div className="upload-options">
                                            <span>options:</span>
                                            <CheckField name="flush_all" checked={this.state.uploadFlushAll}
                                                        onChange={v => this.setState({uploadFlushAll: v})}/>
                                        </div>
                                        <ButtonField variant="green" bordered onClick={this.handleUploadPcap} name="upload" />
                                    </div>

                                    <TextField value={uploadOutput} rows={4} readonly small={true}/>
                                </div>
                            </Col>

                            <Col>
                                <div className="section-header">
                                    <span className="api-request">POST /api/pcap/file</span>
                                    <span className="api-response">{this.state.fileStatusCode}</span>
                                </div>

                                <div className="section-content">
                                    <InputField name="file" active={this.state.isUploadFileFocused}
                                                onChange={this.handleFileChange} value={this.state.uploadSelectedFile}
                                                defaultValue={"local .pcap[ng] path"} inline/>

                                    <div className="upload-actions" style={{"marginTop": "11px"}}>
                                        <div className="upload-options">
                                            <CheckField name="flush_all" checked={this.state.uploadFlushAll}
                                                        onChange={v => this.setState({uploadFlushAll: v})}/>
                                            <CheckField name="delete_original_file" checked={this.state.uploadFlushAll}
                                                        onChange={v => this.setState({uploadFlushAll: v})}/>
                                        </div>
                                        <ButtonField variant="blue" bordered onClick={this.handleUploadPcap} name="process" />
                                    </div>

                                    <TextField value={uploadOutput} rows={4} readonly small={true}/>
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
