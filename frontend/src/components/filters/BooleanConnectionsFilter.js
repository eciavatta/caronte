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
import CheckField from "../fields/CheckField";

class BooleanConnectionsFilter extends Component {

    constructor(props) {
        super(props);
        this.state = {
            filterActive: "false"
        };

        this.filterChanged = this.filterChanged.bind(this);
        this.needRedirect = false;
    }

    componentDidMount() {
        let params = new URLSearchParams(this.props.location.search);
        this.setState({filterActive: this.toBoolean(params.get(this.props.filterName)).toString()});
    }

    componentDidUpdate(prevProps, prevState, snapshot) {
        let urlParams = new URLSearchParams(this.props.location.search);
        let externalActive = this.toBoolean(urlParams.get(this.props.filterName));
        let filterActive = this.toBoolean(this.state.filterActive);
        // if the filterActive state is changed by another component (and not by filterChanged func) and
        // the query string is not equals at the filterActive state, update the state of the component
        if (this.toBoolean(prevState.filterActive) === filterActive && filterActive !== externalActive) {
            this.setState({filterActive: externalActive.toString()});
        }
    }

    toBoolean(value) {
        return value !== null && value.toLowerCase() === "true";
    }

    filterChanged() {
        this.needRedirect = true;
        this.setState({filterActive: (!this.toBoolean(this.state.filterActive)).toString()});
    }

    render() {
        let redirect = null;
        if (this.needRedirect) {
            let urlParams = new URLSearchParams(this.props.location.search);
            if (this.toBoolean(this.state.filterActive)) {
                urlParams.set(this.props.filterName, "true");
            } else {
                urlParams.delete(this.props.filterName);
            }
            redirect = <Redirect push to={`${this.props.location.pathname}?${urlParams}`} />;

            this.needRedirect = false;
        }

        return (
            <div className="filter" style={{"width": `${this.props.width}px`}}>
                <CheckField checked={this.toBoolean(this.state.filterActive)} name={this.props.filterName}
                            onChange={this.filterChanged} />
                {redirect}
            </div>
        );
    }

}

export default withRouter(BooleanConnectionsFilter);
