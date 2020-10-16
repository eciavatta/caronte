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
import Typed from "typed.js";
import dispatcher from "../../dispatcher";
import "./common.scss";
import "./MainPane.scss";
import PcapsPane from "./PcapsPane";
import RulesPane from "./RulesPane";
import ServicesPane from "./ServicesPane";
import StreamsPane from "./StreamsPane";

class MainPane extends Component {

    state = {};

    componentDidMount() {
        const nl = "^600\n^400";
        const options = {
            strings: [
                `welcome to caronte!^1000 the current version is ${this.props.version}` + nl +
                "caronte is a network analyzer,^300 it is able to read pcaps and extract connections", // 0
                "the left panel lists all connections that have already been closed" + nl +
                "scrolling up the list will load the most recent connections,^300 downward the oldest ones", // 1
                "by selecting a connection you can view its content,^300 which will be shown in the right panel" + nl +
                "you can choose the display format,^300 or decide to download the connection content", // 2
                "below there is the timeline,^300 which shows the number of connections per minute per service" + nl +
                "you can use the sliding window to move the time range of the connections to be displayed", // 3
                "there are also additional metrics,^300 selectable from the drop-down menu", // 4
                "at the top are the filters,^300 which can be used to select only certain types of connections" + nl +
                "you can choose which filters to display in the top bar from the filters window", // 5
                "in the pcaps panel it is possible to analyze new pcaps,^300 or to see the pcaps already analyzed" + nl +
                "you can load pcaps from your browser,^300 or process pcaps already present on the filesystem", // 6
                "in the rules panel you can see the rules already created,^300 or create new ones" + nl +
                "the rules inserted will be used only to label new connections, not those already analyzed" + nl +
                "a connection is tagged if it meets all the requirements specified by the rule", // 7
                "in the services panel you can assign new services or edit existing ones" + nl +
                "each service is associated with a port number,^300 and will be shown in the connection list", // 8
                "from the configuration panel you can change the settings of the frontend application", // 9
                "that's all! and have fun!" + nl + "created by @eciavatta" // 10
            ],
            typeSpeed: 40,
            cursorChar: "_",
            backSpeed: 5,
            smartBackspace: false,
            backDelay: 1500,
            preStringTyped: (arrayPos) => {
                switch (arrayPos) {
                    case 1:
                        return dispatcher.dispatch("pulse_connections_view", {duration: 12000});
                    case 2:
                        return this.setState({backgroundPane: <StreamsPane/>});
                    case 3:
                        this.setState({backgroundPane: null});
                        return dispatcher.dispatch("pulse_timeline", {duration: 12000});
                    case 6:
                        return this.setState({backgroundPane: <PcapsPane/>});
                    case 7:
                        return this.setState({backgroundPane: <RulesPane/>});
                    case 8:
                        return this.setState({backgroundPane: <ServicesPane/>});
                    case 10:
                        return this.setState({backgroundPane: null});
                    default:
                        return;
                }
            },
        };
        this.typed = new Typed(this.el, options);
    }

    componentWillUnmount() {
        this.typed.destroy();
    }

    render() {
        return (
            <div className="pane-container">
                <div className="main-pane">
                    <div className="pane-section">
                        <div className="tutorial">
                            <span style={{whiteSpace: "pre"}} ref={(el) => {
                                this.el = el;
                            }}/>
                        </div>
                    </div>
                </div>
                <div className="background-pane">
                    {this.state.backgroundPane}
                </div>
            </div>
        );
    }

}

export default MainPane;
