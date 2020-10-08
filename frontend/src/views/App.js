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
import './App.scss';
import Header from "./Header";
import MainPane from "../components/panels/MainPane";
import Timeline from "./Timeline";
import {BrowserRouter as Router} from "react-router-dom";
import Filters from "./Filters";
import ConfigurationPane from "../components/panels/ConfigurationPane";
import Notifications from "../components/Notifications";
import dispatcher from "../dispatcher";

class App extends Component {

    state = {};

    componentDidMount() {
        dispatcher.register("notifications", payload => {
            if (payload.event === "connected") {
                this.setState({
                    connected: true,
                    configured: payload.message["is_configured"]
                });
            }
        });

        setInterval(() => {
            if (document.title.endsWith("❚")) {
                document.title = document.title.slice(0, -1);
            } else {
                document.title += "❚";
            }
        }, 500);
    }

    render() {
        let modal;
        if (this.state.filterWindowOpen && this.state.configured) {
            modal = <Filters onHide={() => this.setState({filterWindowOpen: false})}/>;
        }

        return (
            <div className="main">
                <Notifications/>
                {this.state.connected &&
                    <Router>
                        <div className="main-header">
                            <Header onOpenFilters={() => this.setState({filterWindowOpen: true})}/>
                        </div>
                        <div className="main-content">
                            {this.state.configured ? <MainPane/> :
                                <ConfigurationPane onConfigured={() => this.setState({configured: true})}/>}
                            {modal}
                        </div>
                        <div className="main-footer">
                            {this.state.configured && <Timeline/>}
                        </div>
                    </Router>
                }
            </div>
        );
    }
}

export default App;
