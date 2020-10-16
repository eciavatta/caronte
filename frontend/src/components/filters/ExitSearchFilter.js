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

class ExitSearchFilter extends Component {

    state = {};

    componentDidMount() {
        let params = new URLSearchParams(this.props.location.search);
        this.setState({performedSearch: params.get("performed_search")});

        this.connectionsFiltersCallback = (payload) => {
            if (this.state.performedSearch !== payload["performed_search"]) {
                this.setState({performedSearch: payload["performed_search"]});
            }
        };
        dispatcher.register("connections_filters", this.connectionsFiltersCallback);
    }

    componentWillUnmount() {
        dispatcher.unregister(this.connectionsFiltersCallback);
    }

    render() {
        return (
            <>
                {this.state.performedSearch &&
                <div className="filter" style={{"width": `${this.props.width}px`}}>
                    <CheckField checked={true} name="exit_search" onChange={() =>
                        dispatcher.dispatch("connections_filters", {"performed_search": null})} small/>
                </div>}
            </>
        );
    }

}

export default withRouter(ExitSearchFilter);
