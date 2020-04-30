import React, {Component} from 'react';
import './Connections.scss';
import axios from 'axios'
import Connection from "../components/Connection";
import Table from 'react-bootstrap/Table';
import {Redirect} from 'react-router';

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
        let redirect = ""
        if (this.state.selected) {
            redirect = <Redirect push to={"/connections/" + this.state.selected} />;
        }

        return (

            <div className="connections">
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
                                selected={this.state.selected === c.id}/>
                        )
                    }
                    </tbody>
                </Table>

                {redirect}
            </div>
        );
    }

}


export default Connections;
