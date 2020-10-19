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
import {createCurlCommand} from "../../utils";
import validation from "../../validation";
import ButtonField from "../fields/ButtonField";
import CheckField from "../fields/CheckField";
import InputField from "../fields/InputField";
import TextField from "../fields/TextField";
import Header from "../Header";
import LinkPopover from "../objects/LinkPopover";
import "../panels/common.scss";
import "./ConfigurationPage.scss";

class ConfigurationPage extends Component {

    constructor(props) {
        super(props);
        this.state = {
            settings: {
                "config": {
                    "server_address": "",
                    "flag_regex": "",
                    "auth_required": false
                },
                "accounts": {}
            },
            newUsername: "",
            newPassword: ""
        };
    }

    saveSettings = () => {
        if (this.validateSettings(this.state.settings)) {
            backend.post("/setup", this.state.settings).then((_) => {
                this.props.onConfigured();
            }).catch((res) => {
                this.setState({setupStatusCode: res.status, setupResponse: JSON.stringify(res.json)});
            });
        }
    };

    validateSettings = (settings) => {
        let valid = true;
        if (!validation.isValidAddress(settings.config["server_address"], true)) {
            this.setState({serverAddressError: "invalid ip_address"});
            valid = false;
        }
        if (settings.config["flag_regex"].length < 8) {
            this.setState({flagRegexError: "flag_regex.length < 8"});
            valid = false;
        }

        return valid;
    };

    updateParam = (callback) => {
        callback(this.state.settings);
        this.setState({settings: this.state.settings});
    };

    addAccount = () => {
        if (this.state.newUsername.length !== 0 && this.state.newPassword.length !== 0) {
            const settings = this.state.settings;
            settings.accounts[this.state.newUsername] = this.state.newPassword;

            this.setState({
                newUsername: "",
                newPassword: "",
                settings
            });
        } else {
            this.setState({
                newUsernameActive: this.state.newUsername.length === 0,
                newPasswordActive: this.state.newPassword.length === 0
            });
        }
    };

    render() {
        const settings = this.state.settings;
        const curlCommand = createCurlCommand("/setup", "POST", settings);

        const accounts = Object.entries(settings.accounts).map(([username, password]) =>
            <tr key={username}>
                <td>{username}</td>
                <td><LinkPopover text="******" content={password}/></td>
                <td><ButtonField variant="red" small rounded name="delete"
                                 onClick={() => this.updateParam((s) => delete s.accounts[username])}/></td>
            </tr>).concat(<tr key={"new_account"}>
            <td><InputField value={this.state.newUsername} small active={this.state.newUsernameActive}
                            onChange={(v) => this.setState({newUsername: v})}/></td>
            <td><InputField value={this.state.newPassword} small active={this.state.newPasswordActive}
                            onChange={(v) => this.setState({newPassword: v})}/></td>
            <td><ButtonField variant="green" small rounded name="add" onClick={this.addAccount}/></td>
        </tr>);

        return (
            <div className="page configuration-page">
                <div className="page-header">
                    <Header />
                </div>

                <div className="page-content">
                    <div className="pane-container configuration-pane">
                        <div className="pane-section">
                            <div className="section-header">
                                <span className="api-request">POST /setup</span>
                                <span className="api-response"><LinkPopover text={this.state.setupStatusCode}
                                                                            content={this.state.setupResponse}
                                                                            placement="left"/></span>
                            </div>

                            <div className="section-content">
                                <Container className="p-0">
                                    <Row>
                                        <Col>
                                            <InputField name="server_address" value={settings.config["server_address"]}
                                                        error={this.state.serverAddressError}
                                                        onChange={(v) => this.updateParam((s) => s.config["server_address"] = v)}/>
                                            <InputField name="flag_regex" value={settings.config["flag_regex"]}
                                                        onChange={(v) => this.updateParam((s) => s.config["flag_regex"] = v)}
                                                        error={this.state.flagRegexError}/>
                                            <div style={{"marginTop": "10px"}}>
                                                <CheckField checked={settings.config["auth_required"]} name="auth_required"
                                                            onChange={(v) => this.updateParam((s) => s.config["auth_required"] = v)}/>
                                            </div>

                                        </Col>

                                        <Col>
                                            accounts:
                                            <div className="section-table">
                                                <Table borderless size="sm">
                                                    <thead>
                                                    <tr>
                                                        <th>username</th>
                                                        <th>password</th>
                                                        <th>actions</th>
                                                    </tr>
                                                    </thead>
                                                    <tbody>
                                                    {accounts}
                                                    </tbody>
                                                </Table>
                                            </div>
                                        </Col>
                                    </Row>
                                </Container>

                                <TextField value={curlCommand} rows={4} readonly small={true}/>
                            </div>

                            <div className="section-footer">
                                <ButtonField variant="green" name="save" bordered onClick={this.saveSettings}/>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        );
    }
}

export default ConfigurationPage;
