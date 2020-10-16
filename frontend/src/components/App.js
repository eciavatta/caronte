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
import {BrowserRouter as Router} from "react-router-dom";
import dispatcher from "../dispatcher";
import Notifications from "./Notifications";
import ConfigurationPage from "./pages/ConfigurationPage";
import MainPage from "./pages/MainPage";
import ServiceUnavailablePage from "./pages/ServiceUnavailablePage";

class App extends Component {

    state = {};

    componentDidMount() {
        dispatcher.register("notifications", this.handleNotifications);

        setInterval(() => {
            if (document.title.endsWith("❚")) {
                document.title = document.title.slice(0, -1);
            } else {
                document.title += "❚";
            }
        }, 500);
    }

    componentWillUnmount() {
        dispatcher.unregister(this.handleNotifications);
    }

    handleNotifications = (payload) => {
        if (payload.event === "connected") {
            this.setState({
                connected: true,
                configured: payload.message["is_configured"],
                version: payload.message["version"]
            });
        }
    };

    render() {
        return (
            <Router>
                <Notifications/>
                {this.state.connected ?
                    (this.state.configured ? <MainPage version={this.state.version}/> :
                        <ConfigurationPage onConfigured={() => this.setState({configured: true})}/>) :
                    <ServiceUnavailablePage/>
                }
            </Router>
        );
    }
}

export default App;
