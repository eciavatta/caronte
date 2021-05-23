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
import {createCurlCommand} from "../../utils";
import validation from "../../validation";
import ButtonField from "../fields/ButtonField";
import ColorField from "../fields/extensions/ColorField";
import NumericField from "../fields/extensions/NumericField";
import InputField from "../fields/InputField";
import TextField from "../fields/TextField";
import LinkPopover from "../objects/LinkPopover";
import "./common.scss";
import "./ServicesPane.scss";

const classNames = require("classnames");
const _ = require("lodash");

class ServicesPane extends Component {

    emptyService = {
        "port": 0,
        "name": "",
        "color": "",
        "notes": ""
    };

    state = {
        services: [],
        currentService: this.emptyService,
    };

    componentDidMount() {
        this.reset();
        this.loadServices();

        dispatcher.register("notifications", this.handleNotifications);
        document.title = "caronte:~/services$";
    }

    componentWillUnmount() {
        dispatcher.unregister(this.handleNotifications);
    }

    handleNotifications = (payload) => {
        if (payload.event === "services.edit") {
            this.loadServices();
        }
    };

    loadServices = () => {
        backend.get("/api/services")
            .then((res) => this.setState({services: Object.values(res.json), servicesStatusCode: res.status}))
            .catch((res) => this.setState({servicesStatusCode: res.status, servicesResponse: JSON.stringify(res.json)}));
    };

    updateService = () => {
        const service = this.state.currentService;
        if (this.validateService(service)) {
            backend.put("/api/services", service).then((res) => {
                this.reset();
                this.setState({serviceStatusCode: res.status});
                this.loadServices();
            }).catch((res) => {
                this.setState({serviceStatusCode: res.status, serviceResponse: JSON.stringify(res.json)});
            });
        }
    };

    deleteService = () => {
        const service = this.state.currentService;
        if (this.validateService(service)) {
            backend.delete("/api/services", service).then((res) => {
                this.reset();
                this.setState({serviceStatusCode: res.status});
                this.loadServices();
            }).catch((res) => {
                this.setState({serviceStatusCode: res.status, serviceResponse: JSON.stringify(res.json)});
            });
        }
    };

    validateService = (service) => {
        let valid = true;
        if (!validation.isValidPort(service.port, true)) {
            this.setState({servicePortError: "port < 0 || port > 65565"});
            valid = false;
        }
        if (service.name.length < 3) {
            this.setState({serviceNameError: "name.length < 3"});
            valid = false;
        }
        if (!validation.isValidColor(service.color)) {
            this.setState({serviceColorError: "color is not hexcolor"});
            valid = false;
        }

        return valid;
    };

    reset = () => {
        this.setState({
            isUpdate: false,
            currentService: _.cloneDeep(this.emptyService),
            servicePortError: null,
            serviceNameError: null,
            serviceColorError: null,
            serviceStatusCode: null,
            servicesStatusCode: null,
            serviceResponse: null,
            servicesResponse: null
        });
    };

    updateParam = (callback) => {
        callback(this.state.currentService);
        this.setState({currentService: this.state.currentService});
    };

    render() {
        const isUpdate = this.state.isUpdate;
        const service = this.state.currentService;

        let services = this.state.services.map((s) =>
            <tr key={s.port} onClick={() => {
                this.reset();
                this.setState({isUpdate: true, currentService: _.cloneDeep(s)});
            }} className={classNames("row-small", "row-clickable", {"row-selected": service.port === s.port})}>
                <td>{s["port"]}</td>
                <td>{s["name"]}</td>
                <td><ButtonField name={s["color"]} color={s["color"]} small/></td>
                <td>{s["notes"]}</td>
            </tr>
        );

        const curlCommand = createCurlCommand("/services", "PUT", service);

        return (
            <div className="pane-container service-pane">
                <div className="pane-section services-list">
                    <div className="section-header">
                        <span className="api-request">GET /api/services</span>
                        {this.state.servicesStatusCode &&
                        <span className="api-response"><LinkPopover text={this.state.servicesStatusCode}
                                                                    content={this.state.servicesResponse}
                                                                    placement="left"/></span>}
                    </div>

                    <div className="section-content">
                        <div className="section-table">
                            <Table borderless size="sm">
                                <thead>
                                <tr>
                                    <th>port</th>
                                    <th>name</th>
                                    <th>color</th>
                                    <th>notes</th>
                                </tr>
                                </thead>
                                <tbody>
                                {services}
                                </tbody>
                            </Table>
                        </div>
                    </div>
                </div>

                <div className="pane-section service-edit">
                    <div className="section-header">
                        <span className="api-request">PUT /api/services</span>
                        <span className="api-response"><LinkPopover text={this.state.serviceStatusCode}
                                                                    content={this.state.serviceResponse}
                                                                    placement="left"/></span>
                    </div>

                    <div className="section-content">
                        <Container className="p-0">
                            <Row>
                                <Col>
                                    <NumericField name="port" value={service.port}
                                                  onChange={(v) => this.updateParam((s) => s.port = v)}
                                                  min={0} max={65565} error={this.state.servicePortError}/>
                                    <InputField name="name" value={service.name}
                                                onChange={(v) => this.updateParam((s) => s.name = v)}
                                                error={this.state.serviceNameError}/>
                                    <ColorField value={service.color} error={this.state.serviceColorError}
                                                onChange={(v) => this.updateParam((s) => s.color = v)}/>
                                </Col>

                                <Col>
                                    <TextField name="notes" rows={7} value={service.notes}
                                               onChange={(v) => this.updateParam((s) => s.notes = v)}/>
                                </Col>
                            </Row>
                        </Container>

                        <TextField value={curlCommand} rows={3} readonly small={true}/>
                    </div>

                    <div className="section-footer">
                        {<ButtonField variant="red" name="cancel" bordered onClick={this.reset}/>}
                        {isUpdate && <ButtonField variant="red" name= "delete_service"
                                     bordered onClick={this.deleteService}/>}
                        <ButtonField variant={isUpdate ? "blue" : "green"}
                                     name={isUpdate ? "update_service" : "add_service"}
                                     bordered onClick={this.updateService}/>
                    </div>
                </div>
            </div>
        );
    }

}

export default ServicesPane;
