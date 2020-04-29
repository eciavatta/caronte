import React, {Component} from 'react';
import './Connections.scss';
import axios from 'axios'
import Connection from "../components/Connection";
import {Link} from "react-router-dom";
import Table from 'react-bootstrap/Table';

class Connections extends Component {
    constructor(props) {
        super(props);
        this.state = {
            connections: [],
        };
    }


    componentDidMount() {
        axios.get("/api/connections").then(res => this.setState({connections: res.data}))
    }

    render() {
        let connection = {
            "id": "5dd95ff0fe7ae01ae7f419c2",
            "ip_src": "10.62.82.1",
            "ip_dst": "10.62.82.2",
            "port_src": 59113,
            "port_dst": 23179,
            "started_at": "2019-11-23T16:36:00.1Z",
            "closed_at": "2019-11-23T16:36:00.971Z",
            "client_bytes": 331,
            "server_bytes": 85,
            "client_documents": 1,
            "server_documents": 1,
            "processed_at": "2020-04-21T17:10:29.532Z",
            "matched_rules": [],
            "hidden": false,
            "marked": true,
            "comment": "",
            "service": {
                "port": 23179,
                "name": "kaboom",
                "color": "#3C6D3C",
                "notes": "wdddoddddddw"
            }
        }

        return (
            <div className="connections">
                <Table striped hover>
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
                    <tr>
                        <td>1</td>
                        <td>Mark</td>
                        <td>Otto</td>
                        <td>@mdo</td>
                    </tr>
                    <tr>
                        <td>2</td>
                        <td>Jacob</td>
                        <td>Thornton</td>
                        <td>@fat</td>
                    </tr>
                    <tr>
                        <td>3</td>
                        <td colSpan="2">Larry the Bird</td>
                        <td>@twitter</td>
                    </tr>
                    </tbody>
                </Table>

                {
                    this.state.connections.map(c =>
                        <Link to={"/connection/" + c.id}><Connection data={c} /></Link>
                    )
                }
            </div>
        );
    }

}


export default Connections;
