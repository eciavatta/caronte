import React, {Component} from 'react';
import './common.scss';
import './MainPane.scss';
import Connections from "../../views/Connections";
import ConnectionContent from "../ConnectionContent";
import {Route, Switch, withRouter} from "react-router-dom";
import PcapPane from "./PcapPane";
import backend from "../../backend";
import RulePane from "./RulePane";
import ServicePane from "./ServicePane";
import log from "../../log";

class MainPane extends Component {

    state = {};

    componentDidMount() {
        const match = this.props.location.pathname.match(/^\/connections\/([a-f0-9]{24})$/);
        if (match != null) {
            this.loading = true;
            backend.get(`/api/connections/${match[1]}`)
                .then(res => {
                    this.loading = false;
                    this.setState({selectedConnection: res.json});
                    log.debug(`Initial connection ${match[1]} loaded`);
                })
                .catch(error => log.error("Error loading initial connection", error));
        }
    }

    render() {
        return (
            <div className="main-pane">
                <div className="pane connections-pane">
                    {
                        !this.loading &&
                        <Connections onSelected={(c) => this.setState({selectedConnection: c})}
                                     initialConnection={this.state.selectedConnection}/>
                    }
                </div>
                <div className="pane details-pane">
                    <Switch>
                        <Route path="/pcaps" children={<PcapPane/>}/>
                        <Route path="/rules" children={<RulePane/>}/>
                        <Route path="/services" children={<ServicePane/>}/>
                        <Route exact path="/connections/:id"
                               children={<ConnectionContent connection={this.state.selectedConnection}/>}/>
                        <Route children={<ConnectionContent/>}/>
                    </Switch>
                </div>
            </div>
        );
    }
}

export default withRouter(MainPane);
