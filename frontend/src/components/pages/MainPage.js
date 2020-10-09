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
import './MainPage.scss';
import './common.scss';
import Connections from "../panels/ConnectionsPane";
import StreamsPane from "../panels/StreamsPane";
import {BrowserRouter as Router, Route, Switch} from "react-router-dom";
import Timeline from "../Timeline";
import PcapsPane from "../panels/PcapsPane";
import RulesPane from "../panels/RulesPane";
import ServicesPane from "../panels/ServicesPane";
import Header from "../Header";
import Filters from "../dialogs/Filters";
import MainPane from "../panels/MainPane";

class MainPage extends Component {

    state = {};

    render() {
        let modal;
        if (this.state.filterWindowOpen) {
            modal = <Filters onHide={() => this.setState({filterWindowOpen: false})}/>;
        }

        return (
            <div className="page main-page">
                <Router>
                    <div className="page-header">
                        <Header onOpenFilters={() => this.setState({filterWindowOpen: true})}/>
                    </div>

                    <div className="page-content">
                        <div className="pane connections-pane">
                            <Connections onSelected={(c) => this.setState({selectedConnection: c})}/>
                        </div>
                        <div className="pane details-pane">
                            <Switch>
                                <Route path="/pcaps" children={<PcapsPane/>}/>
                                <Route path="/rules" children={<RulesPane/>}/>
                                <Route path="/services" children={<ServicesPane/>}/>
                                <Route exact path="/connections/:id"
                                       children={<StreamsPane connection={this.state.selectedConnection}/>}/>
                                <Route children={<MainPane version={this.props.version}/>}/>
                            </Switch>
                        </div>

                        {modal}
                    </div>

                    <div className="page-footer">
                        <Timeline/>
                    </div>
                </Router>
            </div>
        );
    }
}

export default MainPage;
