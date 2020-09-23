import React, {Component} from 'react';
import './MainPane.scss';
import Connections from "./Connections";
import ConnectionContent from "../components/ConnectionContent";
import {withRouter} from "react-router-dom";
import PcapPane from "../components/panels/PcapPane";
import backend from "../backend";

class MainPane extends Component {

    constructor(props) {
        super(props);
        this.state = {
            selectedConnection: null
        };
    }

    componentDidMount() {
        if ('id' in this.props.match.params) {
            const id = this.props.match.params.id;
            backend.get(`/api/connections/${id}`).then(res => {
                if (res.status === 200) {
                    this.setState({selectedConnection: res});
                }
            });
        }
    }

    render() {
        return (
            <div className="main-pane">
                <div className="container-fluid">
                    <div className="row">
                        <div className="col-md-6 pane">
                            <Connections onSelected={(c) => this.setState({selectedConnection: c})} />
                        </div>
                        <div className="col-md-6 pl-0 pane">
                            {/*<PcapPane />*/}
                            <ConnectionContent connection={this.state.selectedConnection}/>
                        </div>
                    </div>
                </div>
            </div>
        );
    }
}

export default withRouter(MainPane);
