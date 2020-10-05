import React, {Component} from 'react';
import './Footer.scss';
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
import dispatcher from "../globals";
import {withRouter} from "react-router-dom";
import log from "../log";


class Footer extends Component {

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
        this.loadStatistics(this.state.metric, filteredPort).then(() =>
            log.debug("Statistics loaded after mount"));

        dispatcher.register((payload) => {
            if (payload.actionType === "connections-update") {
                this.setState({
                    selection: new TimeRange(payload.from, payload.to),
                });
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

        let services = (await backend.get("/api/services")).json;
        if (filteredPort && services[filteredPort]) {
            const service = services[filteredPort];
            services = {};
            services[filteredPort] = service;
        }

        const ports = Object.keys(services);
        ports.forEach(s => urlParams.append("ports", s));

        const metrics = (await backend.get("/api/statistics?" + urlParams)).json;
        const series = new TimeSeries({
            name: "statistics",
            columns: ["time"].concat(ports),
            points: metrics.map(m => [new Date(m["range_start"])].concat(ports.map(p => m[metric][p] || 0)))
        });
        this.setState({
            metric,
            series,
            services,
            timeRange: series.range(),
        });
        log.debug(`Loaded statistics for metric "${metric}" for services [${ports}]`);
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
            dispatcher.dispatch({
                actionType: "timeline-update",
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
        return (
            <footer className="footer">
                <div className="time-line">
                    {this.state.series &&
                    <>
                        <Resizable>
                            <ChartContainer timeRange={this.state.timeRange} enableDragZoom={false}
                                            paddingTop={5} utc minDuration={60000}
                                            maxTime={this.state.series.range().end()}
                                            minTime={this.state.series.range().begin()}
                                            paddingLeft={0} paddingRight={0} paddingBottom={0}
                                            enablePanZoom={true}
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
                    </>
                    }
                </div>
            </footer>
        );
    }
}

export default withRouter(Footer);
