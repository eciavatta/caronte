import React, {Component} from 'react';
import './Connections.scss';
import axios from 'axios'
import Connection from "../components/Connection";
import Table from 'react-bootstrap/Table';
import {Redirect} from 'react-router';
import {objectToQueryString} from "../utils";

class Connections extends Component {

    constructor(props) {
        super(props);
        this.state = {
            loading: false,
            connections: [],
            firstConnection: null,
            lastConnection: null,
            showHidden: false
        };

        this.scrollTopThreashold = 0.00001;
        this.scrollBottomThreashold = 0.99999;
        this.maxConnections = 500;
        this.queryLimit = 50;

        this.handleScroll = this.handleScroll.bind(this);
    }
    
    componentDidMount() {
        this.loadConnections({limit: this.queryLimit, hidden: this.state.showHidden});
    }

    handleScroll(e) {
        let relativeScroll = e.currentTarget.scrollTop / (e.currentTarget.scrollHeight - e.currentTarget.clientHeight);
        if (!this.state.loading && relativeScroll > this.scrollBottomThreashold) {
            this.loadConnections({
                from: this.state.lastConnection.id, limit: this.queryLimit,
                hidden: this.state.showHidden
            });
        }
        if (!this.state.loading && relativeScroll < this.scrollTopThreashold) {
            this.loadConnections({
                to: this.state.firstConnection.id, limit: this.queryLimit,
                hidden: this.state.showHidden
            });
        }

    }

    async loadConnections(params) {
        let url = "/api/connections";
        if (params !== undefined) {
            url += "?" + objectToQueryString(params);
        }
        this.setState({loading: true});
        let res = await axios.get(url);

        let connections = this.state.connections;
        let firstConnection = this.state.firstConnection;
        let lastConnection = this.state.lastConnection;
        if (res.data.length > 0) {
            if (params !== undefined && params.from !== undefined) {
                connections = this.state.connections.concat(res.data);
                lastConnection = connections[connections.length - 1];
                if (connections.length > this.maxConnections) {
                    connections = connections.slice(connections.length - this.maxConnections,
                        connections.length - 1);
                    firstConnection = connections[0];
                }
            } else if (params !== undefined && params.to !== undefined) {
                connections = res.data.concat(this.state.connections);
                firstConnection = connections[0];
                if (connections.length > this.maxConnections) {
                    connections = connections.slice(0, this.maxConnections);
                    lastConnection = connections[this.maxConnections - 1];
                }
            } else {
                connections = res.data;
                firstConnection = connections[0];
                lastConnection = connections[connections.length - 1];
            }
        }

        this.setState({
            loading: false,
            connections: connections,
            firstConnection: firstConnection,
            lastConnection: lastConnection
        });
    }


    render() {
        let redirect = "";
        if (this.state.selected) {
            redirect = <Redirect push to={"/connections/" + this.state.selected}/>;
        }

        let loading = null;
        if (this.state.loading) {
            loading = <tr>
                <td colSpan={9}>Loading...</td>
            </tr>;
        }

        return (
            <div className="connections" onScroll={this.handleScroll}>
                <div className="connections-header-padding"/>
                <Table borderless size="sm">
                    <thead>
                    <tr>
                        <th>service</th>
                        <th>srcip</th>
                        <th>dstip</th>
                        <th>srcport</th>
                        <th>dstport</th>
                        <th>duration</th>
                        <th>up</th>
                        <th>down</th>
                        <th>actions</th>
                    </tr>
                    </thead>
                    <tbody>
                    {
                        this.state.connections.map(c =>
                            <Connection key={c.id} data={c} onSelected={() => this.setState({selected: c.id})}
                                        selected={this.state.selected === c.id} onMarked={marked => c.marked = marked}
                                        onEnabled={enabled => c.hidden = !enabled}/>
                        )
                    }
                    {loading}
                    </tbody>
                </Table>

                {redirect}
            </div>
        );
    }

}


export default Connections;
