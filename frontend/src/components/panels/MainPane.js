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

class MainPane extends Component {

    constructor(props) {
        super(props);
        this.state = {
            selectedConnection: null,
            loading: false
        };
    }

    componentDidMount() {
        const match = this.props.location.pathname.match(/^\/connections\/([a-f0-9]{24})$/);
        if (match != null) {
            this.setState({loading: true});
            backend.get(`/api/connections/${match[1]}`)
                .then(res => this.setState({selectedConnection: res.json, loading: false}))
                .catch(error => console.log(error));
        }
    }

    render() {
        return (
            <div className="main-pane">
                <div className="pane connections-pane">
                    {
                        !this.state.loading &&
                        <Connections onSelected={(c) => this.setState({selectedConnection: c})}
                                     initialConnection={this.state.selectedConnection} />
                    }
                </div>
                <div className="pane details-pane">
                    <Switch>
                        <Route path="/pcaps" children={<PcapPane />} />
                        <Route path="/rules" children={<RulePane />} />
                        <Route path="/services" children={<ServicePane />} />
                        <Route exact path="/connections/:id" children={<ConnectionContent connection={this.state.selectedConnection} />} />
                        <Route children={<ConnectionContent />} />
                    </Switch>
                </div>
            </div>
        );
    }
}

export default withRouter(MainPane);
