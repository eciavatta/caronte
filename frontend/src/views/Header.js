import React, {Component} from 'react';
import Typed from 'typed.js';
import './Header.scss';
import {Button} from "react-bootstrap";
import {filtersDefinitions, filtersNames} from "../components/filters/FiltersDefinitions";
import {Link} from "react-router-dom";
import ButtonField from "../components/fields/ButtonField";

class Header extends Component {

    constructor(props) {
        super(props);
        let newState = {};
        filtersNames.forEach(elem => newState[`${elem}_active`] = false);
        this.state = newState;
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
                        <div className="filters-bar">
                            {quickFilters}
                        </div>
                    </div>

                    <div className="col">
                        <div className="header-buttons">
                            <ButtonField variant="pink" onClick={this.props.onOpenFilters} name="filters" bordered />
                            <Link to="/pcaps">
                                <ButtonField variant="purple" name="pcaps" bordered />
                            </Link>
                            <Link to="/rules">
                                <ButtonField variant="deep-purple" name="rules" bordered />
                            </Link>
                            <ButtonField variant="indigo" onClick={this.props.onOpenServices} name="services" bordered />
                            <ButtonField variant="blue" onClick={this.props.onOpenConfig}
								disabled={false} name="config" bordered />
                        </div>
                    </div>
                </div>
            </header>
        );
    }
}

export default Header;
