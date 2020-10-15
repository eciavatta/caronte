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
import './ConnectionsPane.scss';
import Connection from "../objects/Connection";
import Table from 'react-bootstrap/Table';
import {withRouter} from "react-router-dom";
import backend from "../../backend";
import ConnectionMatchedRules from "../objects/ConnectionMatchedRules";
import log from "../../log";
import ButtonField from "../fields/ButtonField";
import dispatcher from "../../dispatcher";
import {Redirect} from "react-router";

const classNames = require('classnames');

class ConnectionsPane extends Component {

    state = {
        loading: false,
        connections: [],
        firstConnection: null,
        lastConnection: null,
    };

    constructor(props) {
        super(props);

        this.scrollTopThreashold = 0.00001;
        this.scrollBottomThreashold = 0.99999;
        this.maxConnections = 200;
        this.queryLimit = 50;
        this.connectionsListRef = React.createRef();
        this.lastScrollPosition = 0;
    }

    componentDidMount() {
        let urlParams = new URLSearchParams(this.props.location.search);
        this.setState({urlParams});

        const additionalParams = {limit: this.queryLimit};

        const match = this.props.location.pathname.match(/^\/connections\/([a-f0-9]{24})$/);
        if (match != null) {
            const id = match[1];
            additionalParams.from = id;
            backend.get(`/api/connections/${id}`)
                .then(res => this.connectionSelected(res.json))
                .catch(error => log.error("Error loading initial connection", error));
        }

        this.loadConnections(additionalParams, urlParams, true).then(() => log.debug("Connections loaded"));

        this.connectionsFiltersCallback = payload => {
            const params = this.state.urlParams;
            const initialParams = params.toString();

            Object.entries(payload).forEach(([key, value]) => {
                if (value == null) {
                    params.delete(key);
                } else if (Array.isArray(value)) {
                    params.delete(key);
                    value.forEach(v => params.append(key, v));
                } else {
                    params.set(key, value);
                }
            });

            if (initialParams === params.toString()) {
                return;
            }

            log.debug("Update following url params:", payload);
            this.queryStringRedirect = true;
            this.setState({urlParams});

            this.loadConnections({limit: this.queryLimit}, urlParams)
                .then(() => log.info("ConnectionsPane reloaded after query string update"));
        };
        dispatcher.register("connections_filters", this.connectionsFiltersCallback);

        this.timelineUpdatesCallback = payload => {
            this.connectionsListRef.current.scrollTop = 0;
            this.loadConnections({
                started_after: Math.round(payload.from.getTime() / 1000),
                started_before: Math.round(payload.to.getTime() / 1000),
                limit: this.maxConnections
            }).then(() => log.info(`Loading connections between ${payload.from} and ${payload.to}`));
        };
        dispatcher.register("timeline_updates", this.timelineUpdatesCallback);

        this.notificationsCallback = payload => {
            if (payload.event === "rules.new" || payload.event === "rules.edit") {
                this.loadRules().then(() => log.debug("Loaded connection rules after notification update"));
            }
            if (payload.event === "services.edit") {
                this.loadServices().then(() => log.debug("Services reloaded after notification update"));
            }
        };
        dispatcher.register("notifications", this.notificationsCallback);

        this.pulseConnectionsViewCallback = payload => {
            this.setState({pulseConnectionsView: true});
            setTimeout(() => this.setState({pulseConnectionsView: false}), payload.duration);
        };
        dispatcher.register("pulse_connections_view", this.pulseConnectionsViewCallback);
    }

    componentWillUnmount() {
        dispatcher.unregister(this.timelineUpdatesCallback);
        dispatcher.unregister(this.notificationsCallback);
        dispatcher.unregister(this.pulseConnectionsViewCallback);
        dispatcher.unregister(this.connectionsFiltersCallback);
    }

    connectionSelected = (c) => {
        this.connectionSelectedRedirect = true;
        this.setState({selected: c.id});
        this.props.onSelected(c);
        log.debug(`Connection ${c.id} selected`);
    };

    handleScroll = (e) => {
        if (this.disableScrollHandler) {
            this.lastScrollPosition = e.currentTarget.scrollTop;
            return;
        }

        let relativeScroll = e.currentTarget.scrollTop / (e.currentTarget.scrollHeight - e.currentTarget.clientHeight);
        if (!this.state.loading && relativeScroll > this.scrollBottomThreashold) {
            this.loadConnections({from: this.state.lastConnection.id, limit: this.queryLimit,})
                .then(() => log.info("Following connections loaded"));
        }
        if (!this.state.loading && relativeScroll < this.scrollTopThreashold) {
            this.loadConnections({to: this.state.firstConnection.id, limit: this.queryLimit,})
                .then(() => log.info("Previous connections loaded"));
            if (this.state.showMoreRecentButton) {
                this.setState({showMoreRecentButton: false});
            }
        } else {
            if (this.lastScrollPosition > e.currentTarget.scrollTop) {
                if (!this.state.showMoreRecentButton) {
                    this.setState({showMoreRecentButton: true});
                }
            } else {
                if (this.state.showMoreRecentButton) {
                    this.setState({showMoreRecentButton: false});
                }
            }
        }
        this.lastScrollPosition = e.currentTarget.scrollTop;
    };

    async loadConnections(additionalParams, initialParams = null, isInitial = false) {
        if (!initialParams) {
            initialParams = this.state.urlParams;
        }
        const urlParams = new URLSearchParams(initialParams.toString());
        for (const [name, value] of Object.entries(additionalParams)) {
            urlParams.set(name, value);
        }

        this.setState({loading: true});
        if (!this.state.rules) {
            await this.loadRules();
        }
        if (!this.state.services) {
            await this.loadServices();
        }

        let res = (await backend.get(`/api/connections?${urlParams}`)).json;

        let connections = this.state.connections;
        let firstConnection = this.state.firstConnection;
        let lastConnection = this.state.lastConnection;

        if (additionalParams !== undefined && additionalParams.from !== undefined && additionalParams.to === undefined) {
            if (res.length > 0) {
                if (!isInitial) {
                    res = res.slice(1);
                }
                connections = this.state.connections.concat(res);
                lastConnection = connections[connections.length - 1];
                if (isInitial) {
                    firstConnection = connections[0];
                }
                if (connections.length > this.maxConnections) {
                    connections = connections.slice(connections.length - this.maxConnections,
                        connections.length - 1);
                    firstConnection = connections[0];
                }
            }
        } else if (additionalParams !== undefined && additionalParams.to !== undefined && additionalParams.from === undefined) {
            if (res.length > 0) {
                connections = res.slice(0, res.length - 1).concat(this.state.connections);
                firstConnection = connections[0];
                if (connections.length > this.maxConnections) {
                    connections = connections.slice(0, this.maxConnections);
                    lastConnection = connections[this.maxConnections - 1];
                }
            }
        } else {
            if (res.length > 0) {
                connections = res;
                firstConnection = connections[0];
                lastConnection = connections[connections.length - 1];
            } else {
                connections = [];
                firstConnection = null;
                lastConnection = null;
            }
        }

        this.setState({
            loading: false,
            connections: connections,
            firstConnection: firstConnection,
            lastConnection: lastConnection
        });

        if (firstConnection != null && lastConnection != null) {
            dispatcher.dispatch("connection_updates", {
                from: new Date(lastConnection["started_at"]),
                to: new Date(firstConnection["started_at"])
            });
        }
    }

    loadRules = async () => {
        return backend.get("/api/rules").then(res => this.setState({rules: res.json}));
    };

    loadServices = async () => {
        return backend.get("/api/services").then(res => this.setState({services: res.json}));
    };

    render() {
        let redirect;
        if (this.connectionSelectedRedirect) {
            redirect = <Redirect push to={`/connections/${this.state.selected}?${this.state.urlParams}`}/>;
            this.connectionSelectedRedirect = false;
        } else if (this.queryStringRedirect) {
            redirect = <Redirect push to={`${this.props.location.pathname}?${this.state.urlParams}`}/>;
            this.queryStringRedirect = false;
        }

        let loading = null;
        if (this.state.loading) {
            loading = <tr>
                <td colSpan={10}>Loading...</td>
            </tr>;
        }

        return (
            <div className="connections-container">
                {this.state.showMoreRecentButton && <div className="most-recent-button">
                    <ButtonField name="most_recent" variant="green" bordered onClick={() => {
                        this.disableScrollHandler = true;
                        this.connectionsListRef.current.scrollTop = 0;
                        this.loadConnections({limit: this.queryLimit})
                            .then(() => {
                                this.disableScrollHandler = false;
                                log.info("Most recent connections loaded");
                            });
                    }}/>
                </div>}

                <div className={classNames("connections", {"connections-pulse": this.state.pulseConnectionsView})}
                     onScroll={this.handleScroll} ref={this.connectionsListRef}>
                    <Table borderless size="sm">
                        <thead>
                        <tr>
                            <th>service</th>
                            <th>srcip</th>
                            <th>srcport</th>
                            <th>dstip</th>
                            <th>dstport</th>
                            <th>started_at</th>
                            <th>duration</th>
                            <th>up</th>
                            <th>down</th>
                            <th>actions</th>
                        </tr>
                        </thead>
                        <tbody>
                        {
                            this.state.connections.flatMap(c => {
                                return [<Connection key={c.id} data={c} onSelected={() => this.connectionSelected(c)}
                                                    selected={this.state.selected === c.id}
                                                    onMarked={marked => c.marked = marked}
                                                    onEnabled={enabled => c.hidden = !enabled}
                                                    services={this.state.services}/>,
                                    c.matched_rules.length > 0 &&
                                    <ConnectionMatchedRules key={c.id + "_m"} matchedRules={c.matched_rules}
                                                            rules={this.state.rules}/>
                                ];
                            })
                        }
                        {loading}
                        </tbody>
                    </Table>

                    {redirect}
                </div>
            </div>
        );
    }

}

export default withRouter(ConnectionsPane);
