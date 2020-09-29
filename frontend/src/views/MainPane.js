import React, {Component} from 'react';
import './MainPane.scss';
import Connections from "./Connections";
import ConnectionContent from "../components/ConnectionContent";
import {Route, Switch, withRouter} from "react-router-dom";
import PcapPane from "../components/panels/PcapPane";
import backend from "../backend";
import RulePane from "../components/panels/RulePane";

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
            backend.getJson(`/api/connections/${match[1]}`)
                .then(res => this.setState({selectedConnection: res.json, loading: false}))
                .catch(error => console.log(error));
        }
    }

    render() {
        return (
            <div className="main-pane">
                <div className="container-fluid">
                    <div className="row">
                        <div className="col-md-6 pane">
                            {
                                !this.state.loading &&
                                    <Connections onSelected={(c) => this.setState({selectedConnection: c})}
                                                 initialConnection={this.state.selectedConnection} />
                            }
                        </div>
                        <div className="col-md-6 pl-0 pane">
                            <Switch>
                                <Route path="/pcaps" children={<PcapPane />} />
                                <Route path="/rules" children={<RulePane />} />
                                <Route exact path="/connections/:id" children={<ConnectionContent connection={this.state.selectedConnection} />} />
                                <Route children={<ConnectionContent />} />
                            </Switch>
                        </div>
                    </div>
                </div>
            </div>
        );
    }
}

export default withRouter(MainPane);
