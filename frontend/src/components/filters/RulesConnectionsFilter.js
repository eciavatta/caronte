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
import backend from "../../backend";
import dispatcher from "../../dispatcher";
import TagField from "../fields/TagField";

const classNames = require("classnames");
const _ = require("lodash");

class RulesConnectionsFilter extends Component {

    state = {
        rules: [],
        activeRules: []
    };

    componentDidMount() {
        const params = new URLSearchParams(this.props.location.search);
        let activeRules = params.getAll("matched_rules") || [];

        backend.get("/api/rules").then((res) => {
            let rules = res.json.flatMap((rule) => rule.enabled ? [{id: rule.id, name: rule.name}] : []);
            activeRules = rules.filter((rule) => activeRules.some((id) => rule.id === id));
            this.setState({rules, activeRules});
        });

        this.connectionsFiltersCallback = (payload) => {
            if ("matched_rules" in payload && !_.isEqual(payload["matched_rules"].sort(), this.state.activeRules.sort())) {
                const newRules = this.state.rules.filter((r) => payload["matched_rules"].includes(r.id));
                this.setState({
                    activeRules: newRules.map((r) => {
                        return {id: r.id, name: r.name};
                    })
                });
            }
        };
        dispatcher.register("connections_filters", this.connectionsFiltersCallback);
    }

    componentWillUnmount() {
        dispatcher.unregister(this.connectionsFiltersCallback);
    }

    onChange = (activeRules) => {
        if (!_.isEqual(activeRules.sort(), this.state.activeRules.sort())) {
            this.setState({activeRules});
            dispatcher.dispatch("connections_filters", {"matched_rules": activeRules.map((r) => r.id)});
        }
    };

    render() {
        return (
            <div
                className={classNames("filter", "d-inline-block", {"filter-active": this.state.filterActive === "true"})}>
                <div className="filter-rules">
                    <TagField tags={this.state.activeRules} onChange={this.onChange}
                              suggestions={_.differenceWith(this.state.rules, this.state.activeRules, _.isEqual)}
                              minQueryLength={0} name="matched_rules" inline small placeholder="rule_name"/>
                </div>
            </div>
        );
    }

}

export default withRouter(RulesConnectionsFilter);
