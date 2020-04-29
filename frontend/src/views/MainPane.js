import React, {Component} from 'react';
import './MainPane.scss';
import Connections from "./Connections";
import ConnectionContent from "../components/ConnectionContent";
import {withRouter} from "react-router-dom";
import axios from 'axios'

class MainPane extends Component {

    constructor(props) {
        super(props);
        this.state = {
            id: null,
        };
    }

    componentDidUpdate() {
        if (this.props.match.params.id !== this.state.id) {
            const id = this.props.match.params.id;
            this.setState({id: id});

            axios.get(`/api/streams/${id}`).then(res => this.setState({connectionContent: res.data}))


        }
    }

    componentDidMount() {
    }

    render() {
        return (
            <div className="main-pane">
                <div className="container-fluid">
                    <div className="row">
                        <div className="col-md-6 pane">
                            <Connections/>
                        </div>
                        <div className="col-md-6 pl-0 pane">
                            <ConnectionContent connectionPayload={this.state.connectionContent}/>
                        </div>
                    </div>
                </div>
            </div>
        );
    }
}

export default withRouter(MainPane);
