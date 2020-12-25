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

import React, {Component} from "react";
import {Col, Container, Row} from "react-bootstrap";
import Table from "react-bootstrap/Table";
import backend from "../../backend";
import dispatcher from "../../dispatcher";
import validation from "../../validation";
import ButtonField from "../fields/ButtonField";
import CheckField from "../fields/CheckField";
import ChoiceField from "../fields/ChoiceField";
import ColorField from "../fields/extensions/ColorField";
import NumericField from "../fields/extensions/NumericField";
import InputField from "../fields/InputField";
import TextField from "../fields/TextField";
import CopyLinkPopover from "../objects/CopyLinkPopover";
import LinkPopover from "../objects/LinkPopover";
import "./common.scss";
import "./RulesPane.scss";

const classNames = require("classnames");
const _ = require("lodash");

class RulesPane extends Component {

    emptyRule = {
        "name": "",
        "color": "",
        "notes": "",
        "enabled": true,
        "patterns": [],
        "filter": {
            "service_port": 0,
            "client_address": "",
            "client_port": 0,
            "min_duration": 0,
            "max_duration": 0,
            "min_bytes": 0,
            "max_bytes": 0
        },
        "version": 0
    };
    emptyPattern = {
        "regex": "",
        "flags": {
            "caseless": false,
            "dot_all": false,
            "multi_line": false,
            "utf_8_mode": false,
            "unicode_property": false
        },
        "min_occurrences": 0,
        "max_occurrences": 0,
        "direction": 0
    };
    state = {
        rules: [],
        newRule: this.emptyRule,
        newPattern: this.emptyPattern
    };

    constructor(props) {
        super(props);

        this.directions = {
            0: "both",
            1: "c->s",
            2: "s->c"
        };
    }

    componentDidMount() {
        this.reset();
        this.loadRules();

        dispatcher.register("notifications", this.handleNotifications);
        document.title = "caronte:~/rules$";
    }

    componentWillUnmount() {
        dispatcher.unregister(this.handleNotifications);
    }

    handleNotifications = (payload) => {
        if (payload.event === "rules.new" || payload.event === "rules.edit") {
            this.loadRules();
        }
    };

    loadRules = () => {
        backend.get("/api/rules").then((res) => this.setState({rules: res.json, rulesStatusCode: res.status}))
            .catch((res) => this.setState({rulesStatusCode: res.status, rulesResponse: JSON.stringify(res.json)}));
    };

    addRule = () => {
        if (this.validateRule(this.state.newRule)) {
            backend.post("/api/rules", this.state.newRule).then((res) => {
                this.reset();
                this.setState({ruleStatusCode: res.status});
                this.loadRules();
            }).catch((res) => {
                this.setState({ruleStatusCode: res.status, ruleResponse: JSON.stringify(res.json)});
            });
        }
    };

    updateRule = () => {
        const rule = this.state.selectedRule;
        if (this.validateRule(rule)) {
            backend.put(`/api/rules/${rule.id}`, rule).then((res) => {
                this.reset();
                this.setState({ruleStatusCode: res.status});
                this.loadRules();
            }).catch((res) => {
                this.setState({ruleStatusCode: res.status, ruleResponse: JSON.stringify(res.json)});
            });
        }
    };

    validateRule = (rule) => {
        let valid = true;
        if (rule.name.length < 3) {
            this.setState({ruleNameError: "name.length < 3"});
            valid = false;
        }
        if (!validation.isValidColor(rule.color)) {
            this.setState({ruleColorError: "color is not hexcolor"});
            valid = false;
        }
        if (!validation.isValidPort(rule.filter["service_port"])) {
            this.setState({ruleServicePortError: "service_port > 65565"});
            valid = false;
        }
        if (!validation.isValidPort(rule.filter["client_port"])) {
            this.setState({ruleClientPortError: "client_port > 65565"});
            valid = false;
        }
        if (!validation.isValidAddress(rule.filter["client_address"])) {
            this.setState({ruleClientAddressError: "client_address is not ip_address"});
            valid = false;
        }
        if (rule.filter["min_duration"] > rule.filter["max_duration"]) {
            this.setState({ruleDurationError: "min_duration > max_dur."});
            valid = false;
        }
        if (rule.filter["min_bytes"] > rule.filter["max_bytes"]) {
            this.setState({ruleBytesError: "min_bytes > max_bytes"});
            valid = false;
        }
        if (rule.patterns.length < 1) {
            this.setState({rulePatternsError: "patterns.length < 1"});
            valid = false;
        }

        return valid;
    };

    reset = () => {
        const newRule = _.cloneDeep(this.emptyRule);
        const newPattern = _.cloneDeep(this.emptyPattern);
        this.setState({
            selectedRule: null,
            newRule,
            selectedPattern: null,
            newPattern,
            patternRegexFocused: false,
            patternOccurrencesFocused: false,
            ruleNameError: null,
            ruleColorError: null,
            ruleServicePortError: null,
            ruleClientPortError: null,
            ruleClientAddressError: null,
            ruleDurationError: null,
            ruleBytesError: null,
            rulePatternsError: null,
            ruleStatusCode: null,
            rulesStatusCode: null,
            ruleResponse: null,
            rulesResponse: null
        });
    };

    updateParam = (callback) => {
        const updatedRule = this.currentRule();
        callback(updatedRule);
        this.setState({newRule: updatedRule});
    };

    currentRule = () => this.state.selectedRule != null ? this.state.selectedRule : this.state.newRule;

    addPattern = (pattern) => {
        if (!this.validatePattern(pattern)) {
            return;
        }

        const newPattern = _.cloneDeep(this.emptyPattern);
        this.currentRule().patterns.push(pattern);
        this.setState({newPattern});
    };

    editPattern = (pattern) => {
        this.setState({
            selectedPattern: pattern
        });
    };

    updatePattern = (pattern) => {
        if (!this.validatePattern(pattern)) {
            return;
        }

        this.setState({
            selectedPattern: null
        });
    };

    validatePattern = (pattern) => {
        let valid = true;
        if (pattern.regex === "") {
            valid = false;
            this.setState({patternRegexFocused: true});
        }
        if (pattern["min_occurrences"] > pattern["max_occurrences"]) {
            valid = false;
            this.setState({patternOccurrencesFocused: true});
        }
        return valid;
    };

    render() {
        const isUpdate = this.state.selectedRule != null;
        const rule = this.currentRule();
        const pattern = this.state.selectedPattern || this.state.newPattern;

        let rules = this.state.rules.map((r) =>
            <tr key={r.id} onClick={() => {
                this.reset();
                this.setState({selectedRule: _.cloneDeep(r)});
            }} className={classNames("row-small", "row-clickable", {"row-selected": rule.id === r.id})}>
                <td><CopyLinkPopover text={r["id"].substring(0, 8)} value={r["id"]}/></td>
                <td>{r["name"]}</td>
                <td><ButtonField name={r["color"]} color={r["color"]} small/></td>
                <td>{r["notes"]}</td>
            </tr>
        );

        let patterns = (this.state.selectedPattern == null && !isUpdate ?
                rule.patterns.concat(this.state.newPattern) :
                rule.patterns
        ).map((p) => p === pattern ?
            <tr key={"new_pattern"}>
                <td style={{"width": "500px"}}>
                    <InputField small active={this.state.patternRegexFocused} value={pattern.regex}
                                onChange={(v) => {
                                    this.updateParam(() => pattern.regex = v);
                                    this.setState({patternRegexFocused: pattern.regex === ""});
                                }}/>
                </td>
                <td><CheckField small checked={pattern.flags["caseless"]}
                                onChange={(v) => this.updateParam(() => pattern.flags["caseless"] = v)}/></td>
                <td><CheckField small checked={pattern.flags["dot_all"]}
                                onChange={(v) => this.updateParam(() => pattern.flags["dot_all"] = v)}/></td>
                <td><CheckField small checked={pattern.flags["multi_line"]}
                                onChange={(v) => this.updateParam(() => pattern.flags["multi_line"] = v)}/></td>
                <td><CheckField small checked={pattern.flags["utf_8_mode"]}
                                onChange={(v) => this.updateParam(() => pattern.flags["utf_8_mode"] = v)}/></td>
                <td><CheckField small checked={pattern.flags["unicode_property"]}
                                onChange={(v) => this.updateParam(() => pattern.flags["unicode_property"] = v)}/></td>
                <td style={{"width": "70px"}}>
                    <NumericField small value={pattern["min_occurrences"]}
                                  active={this.state.patternOccurrencesFocused}
                                  onChange={(v) => this.updateParam(() => pattern["min_occurrences"] = v)}/>
                </td>
                <td style={{"width": "70px"}}>
                    <NumericField small value={pattern["max_occurrences"]}
                                  active={this.state.patternOccurrencesFocused}
                                  onChange={(v) => this.updateParam(() => pattern["max_occurrences"] = v)}/>
                </td>
                <td><ChoiceField inline small keys={[0, 1, 2]} values={["both", "c->s", "s->c"]}
                                 value={this.directions[pattern.direction]}
                                 onChange={(v) => this.updateParam(() => pattern.direction = v)}/></td>
                <td>{this.state.selectedPattern == null ?
                    <ButtonField variant="green" small name="add" inline rounded onClick={() => this.addPattern(p)}/> :
                    <ButtonField variant="green" small name="save" inline rounded
                                 onClick={() => this.updatePattern(p)}/>}
                </td>
            </tr>
            :
            <tr key={"new_pattern"} className="row-small">
                <td>{p.regex}</td>
                <td>{p.flags["caseless"] ? "yes" : "no"}</td>
                <td>{p.flags["dot_all"] ? "yes" : "no"}</td>
                <td>{p.flags["multi_line"] ? "yes" : "no"}</td>
                <td>{p.flags["utf_8_mode"] ? "yes" : "no"}</td>
                <td>{p.flags["unicode_property"] ? "yes" : "no"}</td>
                <td>{p["min_occurrences"]}</td>
                <td>{p["max_occurrences"]}</td>
                <td>{this.directions[p.direction]}</td>
                {!isUpdate && <td><ButtonField variant="blue" small rounded name="edit"
                                               onClick={() => this.editPattern(p)}/></td>}
            </tr>
        );

        return (
            <div className="pane-container rule-pane">
                <div className="pane-section rules-list">
                    <div className="section-header">
                        <span className="api-request">GET /api/rules</span>
                        {this.state.rulesStatusCode &&
                        <span className="api-response"><LinkPopover text={this.state.rulesStatusCode}
                                                                    content={this.state.rulesResponse}
                                                                    placement="left"/></span>}
                    </div>

                    <div className="section-content">
                        <div className="section-table">
                            <Table borderless size="sm">
                                <thead>
                                <tr>
                                    <th>id</th>
                                    <th>name</th>
                                    <th>color</th>
                                    <th>notes</th>
                                </tr>
                                </thead>
                                <tbody>
                                {rules}
                                </tbody>
                            </Table>
                        </div>
                    </div>
                </div>

                <div className="pane-section rule-edit">
                    <div className="section-header">
                        <span className="api-request">
                            {isUpdate ? `PUT /api/rules/${this.state.selectedRule.id}` : "POST /api/rules"}
                        </span>
                        <span className="api-response"><LinkPopover text={this.state.ruleStatusCode}
                                                                    content={this.state.ruleResponse}
                                                                    placement="left"/></span>
                    </div>

                    <div className="section-content">
                        <Container className="p-0">
                            <Row>
                                <Col>
                                    <InputField name="name" inline value={rule.name}
                                                onChange={(v) => this.updateParam((r) => r.name = v)}
                                                error={this.state.ruleNameError}/>
                                    <ColorField inline value={rule.color} error={this.state.ruleColorError}
                                                onChange={(v) => this.updateParam((r) => r.color = v)}/>
                                    <TextField name="notes" rows={2} value={rule.notes}
                                               onChange={(v) => this.updateParam((r) => r.notes = v)}/>
                                </Col>

                                <Col style={{"paddingTop": "6px"}}>
                                    <span>filters:</span>
                                    <NumericField name="service_port" inline value={rule.filter["service_port"]}
                                                  onChange={(v) => this.updateParam((r) => r.filter["service_port"] = v)}
                                                  min={0} max={65565} error={this.state.ruleServicePortError}
                                                  readonly={isUpdate}/>
                                    <NumericField name="client_port" inline value={rule.filter["client_port"]}
                                                  onChange={(v) => this.updateParam((r) => r.filter["client_port"] = v)}
                                                  min={0} max={65565} error={this.state.ruleClientPortError}
                                                  readonly={isUpdate}/>
                                    <InputField name="client_address" value={rule.filter["client_address"]}
                                                error={this.state.ruleClientAddressError} readonly={isUpdate}
                                                onChange={(v) => this.updateParam((r) => r.filter["client_address"] = v)}/>
                                </Col>

                                <Col style={{"paddingTop": "11px"}}>
                                    <NumericField name="min_duration" inline value={rule.filter["min_duration"]}
                                                  error={this.state.ruleDurationError} readonly={isUpdate}
                                                  onChange={(v) => this.updateParam((r) => r.filter["min_duration"] = v)}/>
                                    <NumericField name="max_duration" inline value={rule.filter["max_duration"]}
                                                  error={this.state.ruleDurationError} readonly={isUpdate}
                                                  onChange={(v) => this.updateParam((r) => r.filter["max_duration"] = v)}/>
                                    <NumericField name="min_bytes" inline value={rule.filter["min_bytes"]}
                                                  error={this.state.ruleBytesError} readonly={isUpdate}
                                                  onChange={(v) => this.updateParam((r) => r.filter["min_bytes"] = v)}/>
                                    <NumericField name="max_bytes" inline value={rule.filter["max_bytes"]}
                                                  error={this.state.ruleBytesError} readonly={isUpdate}
                                                  onChange={(v) => this.updateParam((r) => r.filter["max_bytes"] = v)}/>
                                </Col>
                            </Row>
                        </Container>

                        <div className="section-table">
                            <Table borderless size="sm">
                                <thead>
                                <tr>
                                    <th>regex</th>
                                    <th>!Aa</th>
                                    <th>.*</th>
                                    <th>\n+</th>
                                    <th>UTF8</th>
                                    <th>Uni_</th>
                                    <th>min</th>
                                    <th>max</th>
                                    <th>direction</th>
                                    {!isUpdate && <th>actions</th>}
                                </tr>
                                </thead>
                                <tbody>
                                {patterns}
                                </tbody>
                            </Table>
                            {this.state.rulePatternsError != null &&
                            <span className="table-error">error: {this.state.rulePatternsError}</span>}
                        </div>
                    </div>

                    <div className="section-footer">
                        {<ButtonField variant="red" name="cancel" bordered onClick={this.reset}/>}
                        <ButtonField variant={isUpdate ? "blue" : "green"} name={isUpdate ? "update_rule" : "add_rule"}
                                     bordered onClick={isUpdate ? this.updateRule : this.addRule}/>
                    </div>
                </div>
            </div>
        );
    }

}

export default RulesPane;
