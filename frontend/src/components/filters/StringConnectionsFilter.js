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
import InputField from "../fields/InputField";

class StringConnectionsFilter extends Component {

    state = {
        fieldValue: "",
        filterValue: null,
        timeoutHandle: null,
        invalidValue: false
    };

    componentDidMount() {
        let params = new URLSearchParams(this.props.location.search);
        this.updateStateFromFilterValue(params.get(this.props.filterName));

        this.connectionsFiltersCallback = (payload) => {
            const name = this.props.filterName;
            if (name in payload && this.state.filterValue !== payload[name]) {
                this.updateStateFromFilterValue(payload[name]);
            }
        };
        dispatcher.register("connections_filters", this.connectionsFiltersCallback);
    }

    componentWillUnmount() {
        dispatcher.unregister(this.connectionsFiltersCallback);
    }

    updateStateFromFilterValue = (filterValue) => {
        if (filterValue !== null) {
            let fieldValue = filterValue;
            if (typeof this.props.decodeFunc === "function") {
                fieldValue = this.props.decodeFunc(filterValue);
            }
            if (typeof this.props.replaceFunc === "function") {
                fieldValue = this.props.replaceFunc(fieldValue);
            }
            if (this.isValueValid(fieldValue)) {
                this.setState({fieldValue, filterValue});
            } else {
                this.setState({fieldValue, invalidValue: true});
            }
        } else {
            this.setState({fieldValue: "", filterValue: null});
        }
    };

    isValueValid = (value) => {
        return typeof this.props.validateFunc !== "function" ||
            (typeof this.props.validateFunc === "function" && this.props.validateFunc(value));
    };

    changeFilterValue = (value) => {
        const urlParams = {};
        urlParams[this.props.filterName] = value;
        dispatcher.dispatch("connections_filters", urlParams);
    };

    filterChanged = (fieldValue) => {
        if (this.state.timeoutHandle) {
            clearTimeout(this.state.timeoutHandle);
        }

        if (typeof this.props.replaceFunc === "function") {
            fieldValue = this.props.replaceFunc(fieldValue);
        }

        if (fieldValue === "") {
            this.setState({fieldValue: "", filterValue: null, invalidValue: false});
            return this.changeFilterValue(null);
        }


        if (this.isValueValid(fieldValue)) {
            let filterValue = fieldValue;
            if (filterValue !== "" && typeof this.props.encodeFunc === "function") {
                filterValue = this.props.encodeFunc(filterValue);
            }

            this.setState({
                fieldValue,
                timeoutHandle: setTimeout(() => {
                    this.setState({filterValue});
                    this.changeFilterValue(filterValue);
                }, 500),
                invalidValue: false
            });
        } else {
            this.setState({fieldValue, invalidValue: true});
        }
    };

    render() {
        let active = this.state.filterValue !== null;

        return (
            <div className="filter" style={{"width": `${this.props.width}px`}}>
                <InputField active={active} invalid={this.state.invalidValue} name={this.props.filterName}
                            placeholder={this.props.defaultFilterValue} onChange={this.filterChanged}
                            value={this.state.fieldValue} inline={this.props.inline} small={this.props.small}/>
            </div>
        );
    }

}

export default withRouter(StringConnectionsFilter);
