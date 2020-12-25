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
import Table from "react-bootstrap/Table";
import backend from "../../backend";
import dispatcher from "../../dispatcher";
import {formatSize} from "../../utils";
import ButtonField from "../fields/ButtonField";
import CopyLinkPopover from "../objects/CopyLinkPopover";
import LinkPopover from "../objects/LinkPopover";
import "./common.scss";
import "./StatsPane.scss";

class StatsPane extends Component {

    state = {
        rules: []
    };

    componentDidMount() {
        this.loadStats();
        this.loadResourcesStats();
        this.loadRules();
        dispatcher.register("notifications", this.handleNotifications);
        document.title = "caronte:~/stats$";
        this.intervalToken = setInterval(() => this.loadResourcesStats(), 3000);
    }

    componentWillUnmount() {
        dispatcher.unregister(this.handleNotifications);
        clearInterval(this.intervalToken);
    }

    handleNotifications = (payload) => {
        if (payload.event.startsWith("pcap")) {
            this.loadStats();
        } else if (payload.event.startsWith("rules")) {
            this.loadRules();
        }
    };

    loadStats = () => {
        backend.get("/api/statistics/totals")
            .then((res) => this.setState({stats: res.json, statsStatusCode: res.status}))
            .catch((res) => this.setState({
                stats: res.json, statsStatusCode: res.status,
                statsResponse: JSON.stringify(res.json)
            }));
    };

    loadResourcesStats = () => {
        backend.get("/api/resources/system")
            .then((res) => this.setState({resourcesStats: res.json, resourcesStatsStatusCode: res.status}))
            .catch((res) => this.setState({
                resourcesStats: res.json, resourcesStatsStatusCode: res.status,
                resourcesStatsResponse: JSON.stringify(res.json)
            }));
    };

    loadRules = () => {
        backend.get("/api/rules").then((res) => this.setState({rules: res.json}));
    };

    render() {
        const s = this.state.stats;
        const rs = this.state.resourcesStats;

        const ports = s && s["connections_per_service"] ? Object.keys(s["connections_per_service"]) : [];
        let connections = 0, clientBytes = 0, serverBytes = 0, totalBytes = 0, duration = 0;
        let servicesStats = ports.map((port) => {
            connections += s["connections_per_service"][port];
            clientBytes += s["client_bytes_per_service"][port];
            serverBytes += s["server_bytes_per_service"][port];
            totalBytes += s["total_bytes_per_service"][port];
            duration += s["duration_per_service"][port];

            return <tr key={port} className="row-small row-clickable">
                <td>{port}</td>
                <td>{formatSize(s["connections_per_service"][port])}</td>
                <td>{formatSize(s["client_bytes_per_service"][port])}B</td>
                <td>{formatSize(s["server_bytes_per_service"][port])}B</td>
                <td>{formatSize(s["total_bytes_per_service"][port])}B</td>
                <td>{formatSize(s["duration_per_service"][port] / 1000)}s</td>
            </tr>;
        });
        servicesStats.push(<tr key="totals" className="row-small row-clickable font-weight-bold">
            <td>totals</td>
            <td>{formatSize(connections)}</td>
            <td>{formatSize(clientBytes)}B</td>
            <td>{formatSize(serverBytes)}B</td>
            <td>{formatSize(totalBytes)}B</td>
            <td>{formatSize(duration / 1000)}s</td>
        </tr>);

        const rulesStats = this.state.rules.map((r) =>
            <tr key={r.id} className="row-small row-clickable">
                <td><CopyLinkPopover text={r["id"].substring(0, 8)} value={r["id"]}/></td>
                <td>{r["name"]}</td>
                <td><ButtonField name={r["color"]} color={r["color"]} small/></td>
                <td>{formatSize(s && s["matched_rules"] && s["matched_rules"][r.id] ? s["matched_rules"][r.id] : 0)}</td>
            </tr>
        );

        const cpuStats = (rs ? rs["cpu_times"] : []).map((cpu, index) =>
            <tr key={cpu["cpu"]} className="row-small row-clickable">
                <td>{cpu["cpu"]}</td>
                <td>{cpu["user"]}</td>
                <td>{cpu["system"]}</td>
                <td>{cpu["idle"]}</td>
                <td>{cpu["nice"]}</td>
                <td>{cpu["iowait"]}</td>
                <td>{rs["cpu_percents"][index].toFixed(2)} %</td>
            </tr>
        );

        return (
            <div className="pane-container stats-pane">
                <div className="pane-section stats-list">
                    <div className="section-header">
                        <span className="api-request">GET /api/statistics/totals</span>
                        <span className="api-response"><LinkPopover text={this.state.statsStatusCode}
                                                                    content={this.state.statsResponse}
                                                                    placement="left"/></span>
                    </div>

                    <div className="section-content">
                        <div className="section-table">
                            <Table borderless size="sm">
                                <thead>
                                <tr>
                                    <th>service</th>
                                    <th>connections</th>
                                    <th>client_bytes</th>
                                    <th>server_bytes</th>
                                    <th>total_bytes</th>
                                    <th>duration</th>
                                </tr>
                                </thead>
                                <tbody>
                                {servicesStats}
                                </tbody>
                            </Table>
                        </div>

                        <div className="section-table">
                            <Table borderless size="sm">
                                <thead>
                                <tr>
                                    <th>rule_id</th>
                                    <th>rule_name</th>
                                    <th>rule_color</th>
                                    <th>occurrences</th>
                                </tr>
                                </thead>
                                <tbody>
                                {rulesStats}
                                </tbody>
                            </Table>
                        </div>
                    </div>
                </div>

                <div className="pane-section stats-list" style={{"paddingTop": "10px"}}>
                    <div className="section-header">
                        <span className="api-request">GET /api/resources/system</span>
                        <span className="api-response"><LinkPopover text={this.state.resourcesStatsStatusCode}
                                                                    content={this.state.resourcesStatsResponse}
                                                                    placement="left"/></span>
                    </div>

                    <div className="section-content">
                        <div className="section-table">
                            <Table borderless size="sm">
                                <thead>
                                <tr>
                                    <th>type</th>
                                    <th>total</th>
                                    <th>used</th>
                                    <th>free</th>
                                    <th>shared</th>
                                    <th>buff/cache</th>
                                    <th>available</th>
                                </tr>
                                </thead>
                                <tbody>
                                <tr className="row-small row-clickable">
                                    <td>mem</td>
                                    <td>{rs && formatSize(rs["virtual_memory"]["total"])}</td>
                                    <td>{rs && formatSize(rs["virtual_memory"]["used"])}</td>
                                    <td>{rs && formatSize(rs["virtual_memory"]["free"])}</td>
                                    <td>{rs && formatSize(rs["virtual_memory"]["shared"])}</td>
                                    <td>{rs && formatSize(rs["virtual_memory"]["cached"])}</td>
                                    <td>{rs && formatSize(rs["virtual_memory"]["available"])}</td>
                                </tr>
                                <tr className="row-small row-clickable">
                                    <td>swap</td>
                                    <td>{rs && formatSize(rs["virtual_memory"]["swaptotal"])}</td>
                                    <td>{rs && formatSize(rs["virtual_memory"]["swaptotal"])}</td>
                                    <td>{rs && formatSize(rs["virtual_memory"]["swapfree"])}</td>
                                    <td>-</td>
                                    <td>-</td>
                                    <td>-</td>
                                </tr>
                                </tbody>
                            </Table>
                        </div>

                        <div className="section-table">
                            <Table borderless size="sm">
                                <thead>
                                <tr>
                                    <th>cpu</th>
                                    <th>user</th>
                                    <th>system</th>
                                    <th>idle</th>
                                    <th>nice</th>
                                    <th>iowait</th>
                                    <th>used_percent</th>
                                </tr>
                                </thead>
                                <tbody>
                                {cpuStats}
                                </tbody>
                            </Table>
                        </div>

                        <div className="section-table">
                            <Table borderless size="sm">
                                <thead>
                                <tr>
                                    <th>disk_path</th>
                                    <th>fs_type</th>
                                    <th>total</th>
                                    <th>free</th>
                                    <th>used</th>
                                    <th>used_percent</th>
                                </tr>
                                </thead>
                                <tbody>
                                <tr className="row-small row-clickable">
                                    <td>{rs && rs["disk_usage"]["path"]}</td>
                                    <td>{rs && rs["disk_usage"]["fstype"]}</td>
                                    <td>{rs && formatSize(rs["disk_usage"]["total"])}</td>
                                    <td>{rs && formatSize(rs["disk_usage"]["free"])}</td>
                                    <td>{rs && formatSize(rs["disk_usage"]["used"])}</td>
                                    <td>{rs && rs["disk_usage"]["usedPercent"].toFixed(2)} %</td>
                                </tr>
                                </tbody>
                            </Table>
                        </div>
                    </div>
                </div>
            </div>
        );
    }

}

export default StatsPane;
