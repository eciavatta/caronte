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
import {updateParams} from "../../utils";
import ButtonField from "../fields/ButtonField";

class AdvancedFilters extends Component {

    state = {};

    componentDidMount() {
        this.urlParams = new URLSearchParams(this.props.location.search);

        this.connectionsFiltersCallback = (payload) => {
            this.urlParams = updateParams(this.urlParams, payload);
            const active = ["client_address", "client_port", "min_duration", "max_duration", "min_bytes", "max_bytes"]
                .some((f) => this.urlParams.has(f));
            if (this.state.active !== active) {
                this.setState({active});
            }
        };
        dispatcher.register("connections_filters", this.connectionsFiltersCallback);
    }

    componentWillUnmount() {
        dispatcher.unregister(this.connectionsFiltersCallback);
    }

    render() {
        return (
            <ButtonField onClick={this.props.onClick} name="advanced_filters" small active={this.state.active}/>
        );
    }

}

export default withRouter(AdvancedFilters);
