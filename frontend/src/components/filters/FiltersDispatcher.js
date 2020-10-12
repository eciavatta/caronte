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
import {withRouter} from "react-router-dom";
import {Redirect} from "react-router";
import dispatcher from "../../dispatcher";

class FiltersDispatcher extends Component {

    state = {};

    componentDidMount() {
        let params = new URLSearchParams(this.props.location.search);
        this.setState({params});

        dispatcher.register("connections_filters", payload => {
            const params = this.state.params;

            Object.entries(payload).forEach(([key, value]) => {
                if (value == null) {
                    params.delete(key);
                } else {
                    params.set(key, value);
                }
            });

            this.needRedirect = true;
            this.setState({params});
        });
    }

    render() {
        if (this.needRedirect) {
            this.needRedirect = false;
            return <Redirect push to={`${this.props.location.pathname}?${this.state.params}`}/>;
        }

        return null;
    }
}

export default withRouter(FiltersDispatcher);
