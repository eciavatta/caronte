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
import './ConnectionMatchedRules.scss';
import ButtonField from "../fields/ButtonField";

class ConnectionMatchedRules extends Component {

    constructor(props) {
        super(props);
        this.state = {
        };
    }

    render() {
        const matchedRules = this.props.matchedRules.map(mr => {
            const rule = this.props.rules.find(r => r.id === mr);
            return <ButtonField key={mr} onClick={() => this.props.addMatchedRulesFilter(rule.id)} name={rule.name}
                                color={rule.color} small />;
        });

        return (
            <tr className="connection-matches">
                <td className="row-label">matched_rules:</td>
                <td className="rule-buttons" colSpan={9}>{matchedRules}</td>
            </tr>
        );
    }
}

export default ConnectionMatchedRules;
