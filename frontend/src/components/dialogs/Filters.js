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
import {Modal} from "react-bootstrap";
import {cleanNumber, validateIpAddress, validateMin, validatePort} from "../../utils";
import ButtonField from "../fields/ButtonField";
import StringConnectionsFilter from "../filters/StringConnectionsFilter";
import "./Filters.scss";

class Filters extends Component {

    render() {
        return (
            <Modal
                {...this.props}
                show={true}
                size="lg"
                aria-labelledby="filters-dialog"
                centered
            >
                <Modal.Header>
                    <Modal.Title id="filters-dialog">
                        ~/advanced_filters
                    </Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <div className="advanced-filters d-flex">
                        <div className="flex-fill">
                            <StringConnectionsFilter filterName="client_address"
                                                     defaultFilterValue="all_addresses"
                                                     validateFunc={validateIpAddress}
                                                     key="client_address_filter"/>
                            <StringConnectionsFilter filterName="min_duration"
                                                     defaultFilterValue="0"
                                                     replaceFunc={cleanNumber}
                                                     validateFunc={validateMin(0)}
                                                     key="min_duration_filter"/>
                            <StringConnectionsFilter filterName="min_bytes"
                                                     defaultFilterValue="0"
                                                     replaceFunc={cleanNumber}
                                                     validateFunc={validateMin(0)}
                                                     key="min_bytes_filter"/>
                        </div>

                        <div className="flex-fill">
                            <StringConnectionsFilter filterName="client_port"
                                                     defaultFilterValue="all_ports"
                                                     replaceFunc={cleanNumber}
                                                     validateFunc={validatePort}
                                                     key="client_port_filter"/>
                            <StringConnectionsFilter filterName="max_duration"
                                                     defaultFilterValue="∞"
                                                     replaceFunc={cleanNumber}
                                                     key="max_duration_filter"/>
                            <StringConnectionsFilter filterName="max_bytes"
                                                     defaultFilterValue="∞"
                                                     replaceFunc={cleanNumber}
                                                     key="max_bytes_filter"/>
                        </div>
                    </div>
                </Modal.Body>
                <Modal.Footer className="dialog-footer">
                    <ButtonField variant="red" bordered onClick={this.props.onHide} name="close"/>
                </Modal.Footer>
            </Modal>
        );
    }
}

export default Filters;
