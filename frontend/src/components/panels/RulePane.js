import React, {Component} from 'react';
import './RulePane.scss';
import Table from "react-bootstrap/Table";
import {Button, Col, Container, Row} from "react-bootstrap";
import InputField from "../fields/InputField";
import CheckField from "../fields/CheckField";
import TextField from "../fields/TextField";
import backend from "../../backend";
import NumericField from "../fields/extensions/NumericField";
import ColorField from "../fields/extensions/ColorField";
import ChoiceField from "../fields/ChoiceField";
import ButtonField from "../fields/ButtonField";

class RulePane extends Component {

    constructor(props) {
        super(props);

        this.state = {
            rules: [],
        };
    }

    componentDidMount() {
        this.loadRules();
    }

    loadRules = () => {
        backend.getJson("/api/rules").then(res => this.setState({rules: res}));
    };


    render() {
        let rules = this.state.rules.map(r =>
            <tr className="table-row">
                <td>{r["id"].substring(0, 8)}</td>
                <td>{r["name"]}</td>
                <td>{r["notes"]}</td>
                {/*<td>{((new Date(s["completed_at"]) - new Date(s["started_at"])) / 1000).toFixed(3)}s</td>*/}
                {/*<td>{formatSize(s["size"])}</td>*/}
                {/*<td>{s["processed_packets"]}</td>*/}
                {/*<td>{s["invalid_packets"]}</td>*/}
                {/*<td>undefined</td>*/}
                {/*<td className="table-cell-action"><a target="_blank"*/}
                {/*                                     href={"/api/pcap/sessions/" + s["id"] + "/download"}>download</a>*/}
                {/*</td>*/}
            </tr>
        );

        return (
            <div className="pane-container rule-pane">
                <div className="pane-section">
                    <div className="section-header">
                        <span className="api-request">GET /api/rules</span>
                        <span className="api-response">200 OK</span>
                    </div>

                    <div className="section-table">
                        <Table borderless size="sm">
                            <thead>
                            <tr>
                                <th>id</th>
                                <th>name</th>
                                <th>notes</th>
                            </tr>
                            </thead>
                            <tbody>
                            {rules}
                            </tbody>
                        </Table>
                    </div>
                </div>

                <div className="pane-section">
                    <div className="section-header">
                        <span className="api-request">POST /api/rules</span>
                        <span className="api-response"></span>
                    </div>

                    <div className="section-content">
                        <Container className="p-0">
                            <Row>
                                <Col>
                                    <InputField name="name" inline />
                                    <ColorField value={this.state.test1} onChange={(e) => this.setState({test1: e})} inline />
                                    <TextField name="notes" rows={2} />
                                </Col>

                                <Col>
                                    <div >filters:</div>
                                    <NumericField name="service_port" inline value={this.state.test} onChange={(e) => this.setState({test: e})} validate={(e) => e%2 === 0} />

                                    <NumericField name="client_port" inline />
                                    <InputField name="client_address" />
                                </Col>

                                <Col>
                                    <NumericField name="min_duration" inline />
                                    <NumericField name="max_duration" inline />
                                    <NumericField name="min_bytes" inline />
                                    <NumericField name="max_bytes" inline />

                                </Col>
                            </Row>
                        </Container>

                        <div className="post-rules-actions">
                            <label>options:</label>
                            <div className="rules-options">
                                <CheckField name={"enabled"} />
                            </div>

                            <ButtonField variant="blue" name="clear" bordered />
                            <ButtonField variant="green" name="add_rule" bordered />
                        </div>

                        patterns:
                        <div className="section-table">
                            <Table borderless size="sm">
                                <thead>
                                <tr>
                                    <th>regex</th>
                                    <th>Aa</th>
                                    <th>.*</th>
                                    <th>\n+</th>
                                    <th>UTF8</th>
                                    <th>Uni_</th>
                                    <th>min</th>
                                    <th>max</th>
                                    <th>direction</th>
                                    <th>actions</th>
                                </tr>
                                </thead>
                                <tbody>
                                    <tr>
                                        <td style={{"width": "500px"}}><InputField small /></td>
                                        <td><CheckField small /></td>
                                        <td><CheckField small /></td>
                                        <td><CheckField small /></td>
                                        <td><CheckField small /></td>
                                        <td><CheckField small /></td>
                                        <td style={{"width": "70px"}}><NumericField small /></td>
                                        <td style={{"width": "70px"}}><NumericField small /></td>
                                        <td><ChoiceField small keys={[0, 1, 2]} values={["both", "c->s", "s->c"]} value="both" /></td>
                                        <td><Button  variant="green" size="sm">add</Button></td>
                                    </tr>
                                </tbody>
                            </Table>
                        </div>

                        <ButtonField name="add_rule" variant="green" bordered />
                        <br />
                        <ButtonField name="add_rule" small color="red"/>
                        <br />
                        <ButtonField name="add_rule" bordered border={"green"} />
                    </div>
                </div>


            </div>
        );
    }

}

export default RulePane;
