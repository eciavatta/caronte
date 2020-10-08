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

import React, {Component} from 'react';
import './Timeline.scss';
import {
    ChartContainer,
    ChartRow,
    Charts,
    LineChart,
    MultiBrush,
    Resizable,
    styler,
    YAxis
} from "react-timeseries-charts";
import {TimeRange, TimeSeries} from "pondjs";
import backend from "../backend";
import ChoiceField from "../components/fields/ChoiceField";
import {withRouter} from "react-router-dom";
import log from "../log";
import dispatcher from "../dispatcher";

const minutes = 60 * 1000;

class Timeline extends Component {

    state = {
        metric: "connections_per_service"
    };

    constructor() {
        super();

        this.disableTimeSeriesChanges = false;
        this.selectionTimeout = null;
    }

    filteredPort = () => {
        const urlParams = new URLSearchParams(this.props.location.search);
        return urlParams.get("service_port");
    };

    componentDidMount() {
        const filteredPort = this.filteredPort();
        this.setState({filteredPort});
        this.loadStatistics(this.state.metric, filteredPort).then(() => log.debug("Statistics loaded after mount"));

        dispatcher.register("connection_updates", payload => {
            this.setState({
                selection: new TimeRange(payload.from, payload.to),
            });
        });

        dispatcher.register("notifications", payload => {
            if (payload.event === "services.edit") {
                this.loadServices().then(() => log.debug("Services reloaded after notification update"));
            }
        });
    }

    componentDidUpdate(prevProps, prevState, snapshot) {
        const filteredPort = this.filteredPort();
        if (this.state.filteredPort !== filteredPort) {
            this.setState({filteredPort});
            this.loadStatistics(this.state.metric, filteredPort).then(() =>
                log.debug("Statistics reloaded after filtered port changes"));
        }
    }

    loadStatistics = async (metric, filteredPort) => {
        const urlParams = new URLSearchParams();
        urlParams.set("metric", metric);

        let services = await this.loadServices();
        if (filteredPort && services[filteredPort]) {
            const service = services[filteredPort];
            services = {};
            services[filteredPort] = service;
        }

        const ports = Object.keys(services);
        ports.forEach(s => urlParams.append("ports", s));

        const metrics = (await backend.get("/api/statistics?" + urlParams)).json;
        const zeroFilledMetrics = [];
        const toTime = m => new Date(m["range_start"]).getTime();

        if (metrics.length > 0) {
            let i = 0;
            for (let interval = toTime(metrics[0]); interval <= toTime(metrics[metrics.length - 1]); interval += minutes) {
                if (interval === toTime(metrics[i])) {
                    const m = metrics[i++];
                    m["range_start"] = new Date(m["range_start"]);
                    zeroFilledMetrics.push(m);
                } else {
                    const m = {};
                    m["range_start"] = new Date(interval);
                    m[metric] = {};
                    ports.forEach(p => m[metric][p] = 0);
                    zeroFilledMetrics.push(m);
                }
            }
        }

        const series = new TimeSeries({
            name: "statistics",
            columns: ["time"].concat(ports),
            points: zeroFilledMetrics.map(m => [m["range_start"]].concat(ports.map(p => m[metric][p] || 0)))
        });
        const start = series.range().begin();
        const end = series.range().end();
        start.setTime(start.getTime() - minutes);
        end.setTime(end.getTime() + minutes);

        this.setState({
            metric,
            series,
            timeRange: new TimeRange(start, end),
            start,
            end
        });
        log.debug(`Loaded statistics for metric "${metric}" for services [${ports}]`);
    };

    loadServices = async () => {
        const services = (await backend.get("/api/services")).json;
        this.setState({services});
        return services;
    };

    createStyler = () => {
        return styler(Object.keys(this.state.services).map(port => {
            return {key: port, color: this.state.services[port].color, width: 2};
        }));
    };

    handleTimeRangeChange = (timeRange) => {
        if (!this.disableTimeSeriesChanges) {
            this.setState({timeRange});
        }
    };

    handleSelectionChange = (timeRange) => {
        this.disableTimeSeriesChanges = true;

        this.setState({selection: timeRange});
        if (this.selectionTimeout) {
            clearTimeout(this.selectionTimeout);
        }
        this.selectionTimeout = setTimeout(() => {
            dispatcher.dispatch("timeline_updates", {
                from: timeRange.begin(),
                to: timeRange.end()
            });
            this.selectionTimeout = null;
            this.disableTimeSeriesChanges = false;
        }, 1000);
    };

    aggregateSeries = (func) => {
        const values = this.state.series.columns().map(c => this.state.series[func](c));
        return Math[func](...values);
    };

    render() {
        if (!this.state.series) {
            return null;
        }

        return (
            <footer className="footer">
                <div className="time-line">
                    <Resizable>
                        <ChartContainer timeRange={this.state.timeRange} enableDragZoom={false}
                                        paddingTop={5} minDuration={60000}
                                        maxTime={this.state.end}
                                        minTime={this.state.start}
                                        paddingLeft={0} paddingRight={0} paddingBottom={0}
                                        enablePanZoom={true} utc={false}
                                        onTimeRangeChanged={this.handleTimeRangeChange}>

                            <ChartRow height="125">
                                <YAxis id="axis1" hideAxisLine
                                       min={this.aggregateSeries("min")}
                                       max={this.aggregateSeries("max")} width="35" type="linear" transition={300}/>
                                <Charts>
                                    <LineChart axis="axis1" series={this.state.series}
                                               columns={Object.keys(this.state.services)}
                                               style={this.createStyler()} interpolation="curveBasis"/>

                                    <MultiBrush
                                        timeRanges={[this.state.selection]}
                                        allowSelectionClear={false}
                                        allowFreeDrawing={false}
                                        onTimeRangeChanged={this.handleSelectionChange}
                                    />
                                </Charts>
                            </ChartRow>
                        </ChartContainer>
                    </Resizable>

                    <div className="metric-selection">
                        <ChoiceField inline small
                                     keys={["connections_per_service", "client_bytes_per_service",
                                         "server_bytes_per_service", "duration_per_service"]}
                                     values={["connections_per_service", "client_bytes_per_service",
                                         "server_bytes_per_service", "duration_per_service"]}
                                     onChange={(metric) => this.loadStatistics(metric, this.state.filteredPort)
                                         .then(() => log.debug("Statistics loaded after metric changes"))}
                                     value={this.state.metric}/>
                    </div>
                </div>
            </footer>
        );
    }
}

export default withRouter(Timeline);
