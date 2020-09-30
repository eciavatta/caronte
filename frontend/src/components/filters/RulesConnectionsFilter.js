import React, {Component} from 'react';
import {withRouter} from "react-router-dom";
import {Redirect} from "react-router";
import './RulesConnectionsFilter.scss';
import ReactTags from 'react-tag-autocomplete';
import backend from "../../backend";

const classNames = require('classnames');

class RulesConnectionsFilter extends Component {

    constructor(props) {
        super(props);
        this.state = {
            mounted: false,
            rules: [],
            activeRules: []
        };

        this.needRedirect = false;
    }

    componentDidMount() {
        let params = new URLSearchParams(this.props.location.search);
        let activeRules = params.getAll("matched_rules") || [];

        backend.get("/api/rules").then(res => {
            let rules = res.json.flatMap(rule => rule.enabled ? [{id: rule.id, name: rule.name}] : []);
            activeRules = rules.filter(rule => activeRules.some(id => rule.id === id));
            this.setState({rules, activeRules, mounted: true});
        });
    }

    componentDidUpdate(prevProps, prevState, snapshot) {
        let urlParams = new URLSearchParams(this.props.location.search);
        let externalRules = urlParams.getAll("matched_rules") || [];
        let activeRules = this.state.activeRules.map(r => r.id);
        let compareRules = (first, second) => first.sort().join(",") === second.sort().join(",");
        if (this.state.mounted &&
            compareRules(prevState.activeRules.map(r => r.id), activeRules) &&
            !compareRules(externalRules, activeRules)) {
            this.setState({activeRules: externalRules.map(id => this.state.rules.find(r => r.id === id))});
        }
    }

    onDelete(i) {
        const activeRules = this.state.activeRules.slice(0);
        activeRules.splice(i, 1);
        this.needRedirect = true;
        this.setState({ activeRules });
    }

    onAddition(rule) {
        if (!this.state.activeRules.includes(rule)) {
            const activeRules = [].concat(this.state.activeRules, rule);
            this.needRedirect = true;
            this.setState({activeRules});
        }
    }

    render() {
        let redirect = null;

        if (this.needRedirect) {
            let urlParams = new URLSearchParams(this.props.location.search);
            urlParams.delete("matched_rules");
            this.state.activeRules.forEach(rule => urlParams.append("matched_rules", rule.id));
            redirect = <Redirect push to={`${this.props.location.pathname}?${urlParams}`} />;

            this.needRedirect = false;
        }

        return (
            <div className={classNames("filter", "d-inline-block", {"filter-active" : this.state.filterActive === "true"})}>
                <div className="filter-booleanq">
                    <ReactTags tags={this.state.activeRules} suggestions={this.state.rules}
                        onDelete={this.onDelete.bind(this)} onAddition={this.onAddition.bind(this)}
                               minQueryLength={0} placeholderText="rule_name"
                               suggestionsFilter={(suggestion, query) =>
                                   suggestion.name.startsWith(query) && !this.state.activeRules.includes(suggestion)} />
                </div>

                {redirect}
            </div>
        );
    }

}

export default withRouter(RulesConnectionsFilter);
