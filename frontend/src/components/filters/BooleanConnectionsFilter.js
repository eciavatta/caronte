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
import {withRouter} from "react-router-dom";
import dispatcher from "../../dispatcher";
import CheckField from "../fields/CheckField";

class BooleanConnectionsFilter extends Component {

    state = {
        filterActive: "false"
    };

    componentDidMount() {
        let params = new URLSearchParams(this.props.location.search);
        this.setState({filterActive: this.toBoolean(params.get(this.props.filterName)).toString()});

        this.connectionsFiltersCallback = (payload) => {
            const name = this.props.filterName;
            if (name in payload && this.state.filterActive !== payload[name]) {
                this.setState({filterActive: payload[name]});
            }
        };
        dispatcher.register("connections_filters", this.connectionsFiltersCallback);
    }

    componentWillUnmount() {
        dispatcher.unregister(this.connectionsFiltersCallback);
    }

    toBoolean = (value) => {
        return value !== null && value.toLowerCase() === "true";
    };

    filterChanged = () => {
        const newValue = (!this.toBoolean(this.state.filterActive)).toString();
        const urlParams = {};
        urlParams[this.props.filterName] = newValue === "true" ? "true" : null;
        dispatcher.dispatch("connections_filters", urlParams);
        this.setState({filterActive: newValue});
    };

    render() {
        return (
            <div className="filter" style={{"width": `${this.props.width}px`}}>
                <CheckField checked={this.toBoolean(this.state.filterActive)} name={this.props.filterName}
                            onChange={this.filterChanged} small/>
            </div>
        );
    }

}

export default withRouter(BooleanConnectionsFilter);
