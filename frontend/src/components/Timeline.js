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

import {TimeRange, TimeSeries} from "pondjs";
import React, {Component} from "react";
import {withRouter} from "react-router-dom";
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
import backend from "../backend";
import dispatcher from "../dispatcher";
import log from "../log";
import ChoiceField from "./fields/ChoiceField";
import "./Timeline.scss";

const minutes = 60 * 1000;
const maxTimelineRange = 24 * 60 * minutes;
const classNames = require("classnames");

const leftSelectionPaddingMultiplier = 24;
const rightSelectionPaddingMultiplier = 8;

class Timeline extends Component {

    state = {
        metric: "connections_per_service"
    };

    constructor() {
        super();

        this.disableTimeSeriesChanges = false;
        this.selectionTimeout = null;
    }

    componentDidMount() {
        const urlParams = new URLSearchParams(this.props.location.search);
        this.setState({
            servicePortFilter: urlParams.get("service_port") || null,
            matchedRulesFilter: urlParams.getAll("matched_rules") || null
        });

        this.loadStatistics(this.state.metric).then(() => log.debug("Statistics loaded after mount"));
        dispatcher.register("connections_filters", this.handleConnectionsFiltersCallback);
        dispatcher.register("connection_updates", this.handleConnectionUpdates);
        dispatcher.register("notifications", this.handleNotifications);
        dispatcher.register("pulse_timeline", this.handlePulseTimeline);
    }

    componentWillUnmount() {
        dispatcher.unregister(this.handleConnectionsFiltersCallback);
        dispatcher.unregister(this.handleConnectionUpdates);
        dispatcher.unregister(this.handleNotifications);
        dispatcher.unregister(this.handlePulseTimeline);
    }

    loadStatistics = async (metric) => {
        const urlParams = new URLSearchParams();
        urlParams.set("metric", metric);

        let columns = [];
        if (metric === "matched_rules") {
            let rules = await this.loadRules();
            if (this.state.matchedRulesFilter.length > 0) {
                this.state.matchedRulesFilter.forEach((id) => {
                    urlParams.append("rules_ids", id);
                });
                columns = this.state.matchedRulesFilter;
            } else {
                columns = rules.map((r) => r.id);
            }
        } else {
            let services = await this.loadServices();
            const filteredPort = this.state.servicePortFilter;
            if (filteredPort && services[filteredPort]) {
                const service = services[filteredPort];
                services = {};
                services[filteredPort] = service;
            }

            columns = Object.keys(services);
            columns.forEach((port) => urlParams.append("ports", port));
        }

        const metrics = (await backend.get("/api/statistics?" + urlParams)).json;
        if (metrics.length === 0) {
            return;
        }

        const zeroFilledMetrics = [];
        const toTime = (m) => new Date(m["range_start"]).getTime();

        let i;
        let timeStart = toTime(metrics[0]) - minutes;
        for (i = 0; timeStart < 0 && i < metrics.length; i++) { // workaround to remove negative timestamps :(
            timeStart = toTime(metrics[i]) - minutes;
        }

        let timeEnd = toTime(metrics[metrics.length - 1]) + minutes;
        if (timeEnd - timeStart > maxTimelineRange) {
            timeEnd = timeStart + maxTimelineRange;

            const now = new Date().getTime();
            if (!this.lastDisplayNotificationTime || this.lastDisplayNotificationTime + minutes < now) {
                this.lastDisplayNotificationTime = now;
                dispatcher.dispatch("notifications", {event: "timeline.range.large"});
            }
        }

        for (let interval = timeStart; interval <= timeEnd; interval += minutes) {
            if (i < metrics.length && interval === toTime(metrics[i])) {
                const m = metrics[i++];
                m["range_start"] = new Date(m["range_start"]);
                zeroFilledMetrics.push(m);
            } else {
                const m = {};
                m["range_start"] = new Date(interval);
                m[metric] = {};
                columns.forEach((c) => m[metric][c] = 0);
                zeroFilledMetrics.push(m);
            }
        }

        const series = new TimeSeries({
            name: "statistics",
            columns: ["time"].concat(columns),
            points: zeroFilledMetrics.map((m) => [m["range_start"]].concat(columns.map((c) =>
                ((metric in m) && (m[metric] != null)) ? (m[metric][c] || 0) : 0
            )))
        });

        const start = series.range().begin();
        const end = series.range().end();

        this.setState({
            metric,
            series,
            timeRange: new TimeRange(start, end),
            columns,
            start,
            end
        });
    };

    loadServices = async () => {
        const services = (await backend.get("/api/services")).json;
        this.setState({services});
        return services;
    };

    loadRules = async () => {
        const rules = (await backend.get("/api/rules")).json;
        this.setState({rules});
        return rules;
    };

    createStyler = () => {
        if (this.state.metric === "matched_rules") {
            return styler(this.state.rules.map((rule) => {
                return {key: rule.id, color: rule.color, width: 2};
            }));
        } else {
            return styler(Object.keys(this.state.services).map((port) => {
                return {key: port, color: this.state.services[port].color, width: 2};
            }));
        }
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

    handleConnectionsFiltersCallback = (payload) => {
        if ("service_port" in payload && this.state.servicePortFilter !== payload["service_port"]) {
            this.setState({servicePortFilter: payload["service_port"]});
            this.loadStatistics(this.state.metric).then(() => log.debug("Statistics reloaded after service port changed"));
        }
        if ("matched_rules" in payload && this.state.matchedRulesFilter !== payload["matched_rules"]) {
            this.setState({matchedRulesFilter: payload["matched_rules"]});
            this.loadStatistics(this.state.metric).then(() => log.debug("Statistics reloaded after matched rules changed"));
        }
    };

    handleConnectionUpdates = (payload) => {
        if (payload.from >= this.state.start && payload.from < payload.to && payload.to <= this.state.end) {
            this.setState({
                selection: new TimeRange(payload.from, payload.to),
            });
            this.adjustSelection();
        }
    };

    handleNotifications = (payload) => {
        if (payload.event === "services.edit" && this.state.metric !== "matched_rules") {
            this.loadStatistics(this.state.metric).then(() => log.debug("Statistics reloaded after services updates"));
        } else if (payload.event.startsWith("rules") && this.state.metric === "matched_rules") {
            this.loadStatistics(this.state.metric).then(() => log.debug("Statistics reloaded after rules updates"));
        } else if (payload.event === "pcap.completed") {
            this.loadStatistics(this.state.metric).then(() => log.debug("Statistics reloaded after pcap processed"));
        }
    };

    handlePulseTimeline = (payload) => {
        this.setState({pulseTimeline: true});
        setTimeout(() => this.setState({pulseTimeline: false}), payload.duration);
    };

    adjustSelection = () => {
        const seriesRange = this.state.series.range();
        const selection = this.state.selection;
        const delta = selection.end() - selection.begin();
        const start = Math.max(selection.begin().getTime() - delta * leftSelectionPaddingMultiplier, seriesRange.begin().getTime());
        const end = Math.min(selection.end().getTime() + delta * rightSelectionPaddingMultiplier, seriesRange.end().getTime());
        this.setState({timeRange: new TimeRange(start, end)});
    };

    aggregateSeries = (func) => {
        const values = this.state.series.columns().map((c) => this.state.series[func](c));
        return Math[func](...values);
    };

    render() {
        if (!this.state.series) {
            return null;
        }

        return (
            <footer className="footer">
                <div className={classNames("time-line", {"pulse-timeline": this.state.pulseTimeline})}>
                    <Resizable>
                        <ChartContainer timeRange={this.state.timeRange} enableDragZoom={false}
                                        paddingTop={5} minDuration={60000}
                                        maxTime={this.state.end}
                                        minTime={this.state.start}
                                        paddingLeft={0} paddingRight={0} paddingBottom={0}
                                        enablePanZoom={true} utc={false}
                                        onTimeRangeChanged={this.handleTimeRangeChange}>

                            <ChartRow height={this.props.height - 70}>
                                <YAxis id="axis1" hideAxisLine
                                       min={this.aggregateSeries("min")}
                                       max={this.aggregateSeries("max")} width="35" type="linear" transition={300}/>
                                <Charts>
                                    <LineChart axis="axis1" series={this.state.series}
                                               columns={this.state.columns}
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
                                         "server_bytes_per_service", "duration_per_service", "matched_rules"]}
                                     values={["connections_per_service", "client_bytes_per_service",
                                         "server_bytes_per_service", "duration_per_service", "matched_rules"]}
                                     onChange={(metric) => this.loadStatistics(metric)
                                         .then(() => log.debug("Statistics loaded after metric changes"))}
                                     value={this.state.metric}/>
                    </div>
                </div>
            </footer>
        );
    }
}

export default withRouter(Timeline);
