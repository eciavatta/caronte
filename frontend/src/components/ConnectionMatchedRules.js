import React, {Component} from 'react';
import './ConnectionMatchedRules.scss';
import ButtonField from "./fields/ButtonField";

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
