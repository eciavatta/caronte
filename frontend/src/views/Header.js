import React, {Component} from 'react';
import Typed from 'typed.js';
import './Header.scss';
import {Button} from "react-bootstrap";
import StringConnectionsFilter from "../components/filters/StringConnectionsFilter";
import {cleanNumber, validateIpAddress, validateMin, validatePort} from "../utils";
import RulesConnectionsFilter from "../components/filters/RulesConnectionsFilter";
import {filtersDefinitions, filtersNames} from "../components/filters/FiltersDefinitions";

class Header extends Component {

    constructor(props) {
        super(props);
        this.state = {};
        filtersNames.forEach(elem => this.state[`${elem}_active`] = false);
        this.fetchStateFromLocalStorage = this.fetchStateFromLocalStorage.bind(this);
    }

    componentDidMount() {
        const options = {
            strings: ["caronte$ "],
            typeSpeed: 50,
            cursorChar: "❚"
        };
        this.typed = new Typed(this.el, options);

        this.fetchStateFromLocalStorage();

        if (typeof window !== "undefined") {
            window.addEventListener("quick-filters", this.fetchStateFromLocalStorage);
        }
    }

    componentWillUnmount() {
        this.typed.destroy();

        if (typeof window !== "undefined") {
            window.removeEventListener("quick-filters", this.fetchStateFromLocalStorage);
        }
    }

    fetchStateFromLocalStorage() {
        let newState = {};
        filtersNames.forEach(elem => newState[`${elem}_active`] = localStorage.getItem(`filters.${elem}`) === "true");
        this.setState(newState);
    }

    render() {
        let quickFilters = filtersNames.filter(name => this.state[`${name}_active`])
            .map(name => filtersDefinitions[name])
            .slice(0, 5);

        return (
            <header className="header container-fluid">
                <div className="row">
                    <div className="col-auto">
                        <h1 className="header-title type-wrap">
                            <span style={{whiteSpace: 'pre'}} ref={(el) => {
                                this.el = el;
                            }}/>
                        </h1>
                    </div>

                    <div className="col-auto">
                        <div className="filters-bar-wrapper">
                            <div className="filters-bar">
                                {quickFilters}
                            </div>
                        </div>
                    </div>

                    <div className="col">
                        <div className="header-buttons">
                            <Button onClick={this.props.onOpenFilters}>filters</Button>
                            <Button variant="yellow" size="sm">pcaps</Button>
                            <Button variant="blue">rules</Button>
                            <Button variant="red" onClick={this.props.onOpenServices}>
                                services
                            </Button>
                        </div>
                    </div>
                </div>
            </header>
        );
    }
}

export default Header;
