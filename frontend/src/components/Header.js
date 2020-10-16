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
import {Link, withRouter} from "react-router-dom";
import Typed from "typed.js";
import {cleanNumber, validatePort} from "../utils";
import ButtonField from "./fields/ButtonField";
import AdvancedFilters from "./filters/AdvancedFilters";
import BooleanConnectionsFilter from "./filters/BooleanConnectionsFilter";
import ExitSearchFilter from "./filters/ExitSearchFilter";
import RulesConnectionsFilter from "./filters/RulesConnectionsFilter";
import StringConnectionsFilter from "./filters/StringConnectionsFilter";
import "./Header.scss";

class Header extends Component {

    componentDidMount() {
        const options = {
            strings: ["caronte$ "],
            typeSpeed: 50,
            cursorChar: "❚"
        };
        this.typed = new Typed(this.el, options);
    }

    componentWillUnmount() {
        this.typed.destroy();
    }

    render() {
        return (
            <header className="header container-fluid">
                <div className="row">
                    <div className="col-auto">
                        <h1 className="header-title type-wrap">
                            <Link to="/">
                                <span style={{whiteSpace: "pre"}} ref={(el) => {
                                    this.el = el;
                                }}/>
                            </Link>
                        </h1>
                    </div>

                    <div className="col-auto">
                        <div className="filters-bar">
                            <StringConnectionsFilter filterName="service_port"
                                                     defaultFilterValue="all_ports"
                                                     replaceFunc={cleanNumber}
                                                     validateFunc={validatePort}
                                                     key="service_port_filter"
                                                     width={200} small inline/>
                            <RulesConnectionsFilter/>
                            <BooleanConnectionsFilter filterName={"marked"}/>
                            <ExitSearchFilter/>
                            <AdvancedFilters onClick={this.props.onOpenFilters}/>
                        </div>
                    </div>

                    <div className="col">
                        <div className="header-buttons">
                            <Link to={"/searches" + this.props.location.search}>
                                <ButtonField variant="pink" name="searches" bordered/>
                            </Link>
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
