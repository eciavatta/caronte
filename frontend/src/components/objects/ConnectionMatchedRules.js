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
import ButtonField from "../fields/ButtonField";
import "./ConnectionMatchedRules.scss";

class ConnectionMatchedRules extends Component {

    onMatchedRulesSelected = (id) => {
        const params = new URLSearchParams(this.props.location.search);
        const rules = params.getAll("matched_rules");
        if (!rules.includes(id)) {
            rules.push(id);
            dispatcher.dispatch("connections_filters", {"matched_rules": rules});
        }
    };

    render() {
        const matchedRules = this.props.matchedRules.map((mr) => {
            const rule = this.props.rules.find((r) => r.id === mr);
            return <ButtonField key={mr} onClick={() => this.onMatchedRulesSelected(rule.id)} name={rule.name}
                                color={rule.color} small/>;
        });

        return (
            <tr className="connection-matches">
                <td className="row-label">matched_rules:</td>
                <td className="rule-buttons" colSpan={9}>{matchedRules}</td>
            </tr>
        );
    }
}

export default withRouter(ConnectionMatchedRules);
