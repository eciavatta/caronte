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

import React, {Component} from "react";
import {Route, Switch} from "react-router-dom";
import Filters from "../dialogs/Filters";
import Header from "../Header";
import Connections from "../panels/ConnectionsPane";
import MainPane from "../panels/MainPane";
import PcapsPane from "../panels/PcapsPane";
import RulesPane from "../panels/RulesPane";
import SearchPane from "../panels/SearchPane";
import ServicesPane from "../panels/ServicesPane";
import StreamsPane from "../panels/StreamsPane";
import Timeline from "../Timeline";
import "./common.scss";
import "./MainPage.scss";

class MainPage extends Component {

    state = {};

    render() {
        let modal;
        if (this.state.filterWindowOpen) {
            modal = <Filters onHide={() => this.setState({filterWindowOpen: false})}/>;
        }

        return (
            <div className="page main-page">
                <div className="page-header">
                    <Header onOpenFilters={() => this.setState({filterWindowOpen: true})} configured={true}/>
                </div>

                <div className="page-content">
                    <div className="pane connections-pane">
                        <Connections onSelected={(c) => this.setState({selectedConnection: c})}/>
                    </div>
                    <div className="pane details-pane">
                        <Switch>
                            <Route path="/searches" children={<SearchPane/>}/>
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
            </div>
        );
    }
}

export default MainPage;
