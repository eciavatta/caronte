import React, {Component} from 'react';
import './Connections.scss';
import Connection from "../components/Connection";
import Table from 'react-bootstrap/Table';
import {Redirect} from 'react-router';
import {withRouter} from "react-router-dom";
import backend from "../backend";
import ConnectionMatchedRules from "../components/ConnectionMatchedRules";
import log from "../log";
import ButtonField from "../components/fields/ButtonField";
import dispatcher from "../dispatcher";

class Connections extends Component {

    state = {
        loading: false,
        connections: [],
        firstConnection: null,
        lastConnection: null,
        queryString: null
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
        this.loadConnections({limit: this.queryLimit})
            .then(() => this.setState({loaded: true}));
        if (this.props.initialConnection != null) {
            this.setState({selected: this.props.initialConnection.id});
            // TODO: scroll to initial connection
        }

        dispatcher.register("timeline_updates", payload => {
            this.connectionsListRef.current.scrollTop = 0;
            this.loadConnections({
                started_after: Math.round(payload.from.getTime() / 1000),
                started_before: Math.round(payload.to.getTime() / 1000),
                limit: this.maxConnections
            }).then(() => log.info(`Loading connections between ${payload.from} and ${payload.to}`));
        });

        dispatcher.register("notifications", payload => {
            if (payload.event === "rules.new" || payload.event === "rules.edit") {
                this.loadRules().then(() => log.debug("Loaded connection rules after notification update"));
            }
        });

        dispatcher.register("notifications", payload => {
            if (payload.event === "services.edit") {
                this.loadServices().then(() => log.debug("Services reloaded after notification update"));
            }
        });
    }

    connectionSelected = (c) => {
        this.setState({selected: c.id});
        this.props.onSelected(c);
    };

    componentDidUpdate(prevProps, prevState, snapshot) {
        if (this.state.loaded && prevProps.location.search !== this.props.location.search) {
            this.setState({queryString: this.props.location.search});
            this.loadConnections({limit: this.queryLimit})
                .then(() => log.info("Connections reloaded after query string update"));
        }
    }

    handleScroll = (e) => {
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

    addServicePortFilter = (port) => {
        const urlParams = new URLSearchParams(this.props.location.search);
        urlParams.set("service_port", port);
        this.setState({queryString: "?" + urlParams});
    };

    addMatchedRulesFilter = (matchedRule) => {
        const urlParams = new URLSearchParams(this.props.location.search);
        const oldMatchedRules = urlParams.getAll("matched_rules") || [];

        if (!oldMatchedRules.includes(matchedRule)) {
            urlParams.append("matched_rules", matchedRule);
            this.setState({queryString: "?" + urlParams});
        }
    };

    async loadConnections(params) {
        let url = "/api/connections";
        const urlParams = new URLSearchParams(this.props.location.search);
        for (const [name, value] of Object.entries(params)) {
            urlParams.set(name, value);
        }

        this.setState({loading: true});
        if (!this.state.rules) {
            await this.loadRules();
        }
        if (!this.state.services) {
            await this.loadServices();
        }

        let res = (await backend.get(`${url}?${urlParams}`)).json;

        let connections = this.state.connections;
        let firstConnection = this.state.firstConnection;
        let lastConnection = this.state.lastConnection;

        if (params !== undefined && params.from !== undefined && params.to === undefined) {
            if (res.length > 0) {
                connections = this.state.connections.concat(res);
                lastConnection = connections[connections.length - 1];
                if (connections.length > this.maxConnections) {
                    connections = connections.slice(connections.length - this.maxConnections,
                        connections.length - 1);
                    firstConnection = connections[0];
                }
            }
        } else if (params !== undefined && params.to !== undefined && params.from === undefined) {
            if (res.length > 0) {
                connections = res.concat(this.state.connections);
                firstConnection = connections[0];
                if (connections.length > this.maxConnections) {
                    connections = connections.slice(0, this.maxConnections);
                    lastConnection = connections[this.maxConnections - 1];
                }
            }
        } else {
            this.connectionsListRef.current.scrollTop = 0;
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
        let queryString = this.state.queryString !== null ? this.state.queryString : "";
        if (this.state.selected) {
            redirect = <Redirect push to={`/connections/${this.state.selected}${queryString}`}/>;
        }

        let loading = null;
        if (this.state.loading) {
            loading = <tr>
                <td colSpan={9}>Loading...</td>
            </tr>;
        }

        return (
            <div className="connections-container">
                {this.state.showMoreRecentButton && <div className="most-recent-button">
                    <ButtonField name="most_recent" variant="green" onClick={() =>
                        this.loadConnections({limit: this.queryLimit})
                            .then(() => log.info("Most recent connections loaded"))
                    }/>
                </div>}

                <div className="connections" onScroll={this.handleScroll} ref={this.connectionsListRef}>
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
                                                    addServicePortFilter={this.addServicePortFilter}
                                                    services={this.state.services}/>,
                                    c.matched_rules.length > 0 &&
                                    <ConnectionMatchedRules key={c.id + "_m"} matchedRules={c.matched_rules}
                                                            rules={this.state.rules}
                                                            addMatchedRulesFilter={this.addMatchedRulesFilter}/>
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

export default withRouter(Connections);
