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
import Typed from 'typed.js';
import './Header.scss';
import {filtersDefinitions, filtersNames} from "../components/filters/FiltersDefinitions";
import {Link, withRouter} from "react-router-dom";
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
            .map(name => <React.Fragment key={name}>{filtersDefinitions[name]}</React.Fragment>)
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
                            <ButtonField variant="pink" onClick={this.props.onOpenFilters} name="filters" bordered/>
                            <Link to={"/pcaps" + this.props.location.search}>
                                <ButtonField variant="purple" name="pcaps" bordered/>
                            </Link>
                            <Link to={"/rules" + this.props.location.search}>
                                <ButtonField variant="deep-purple" name="rules" bordered/>
                            </Link>
                            <Link to={"/services" + this.props.location.search}>
                                <ButtonField variant="indigo" name="services" bordered/>
                            </Link>
                            <Link to={"/config" + this.props.location.search}>
                                <ButtonField variant="blue" name="config" bordered/>
                            </Link>
                        </div>
                    </div>
                </div>
            </header>
        );
    }
}

export default withRouter(Header);
