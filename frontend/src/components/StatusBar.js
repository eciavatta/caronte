/*
 * This file is part of caronte (https://github.com/eciavatta/caronte).
 * Copyright (c) 2021 Emiliano Ciavatta.
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

import React, { Component } from "react";
import { Link, withRouter } from "react-router-dom";
import backend from "../backend";
import dispatcher from "../dispatcher";
import Typed from "typed.js";
import { cleanNumber, validatePort } from "../utils";
import ButtonField from "./fields/ButtonField";
import CheckField from "./fields/CheckField";
import AdvancedFilters from "./filters/AdvancedFilters";
import BooleanConnectionsFilter from "./filters/BooleanConnectionsFilter";
import ExitSearchFilter from "./filters/ExitSearchFilter";
import ExitSimilarityFilter from "./filters/ExitSimilarityFilter";
import RulesConnectionsFilter from "./filters/RulesConnectionsFilter";
import StringConnectionsFilter from "./filters/StringConnectionsFilter";
import LinkPopover from "./objects/LinkPopover";
import "./StatusBar.scss";

const classNames = require("classnames");

class StatusBar extends Component {
  state = {
    status: {},
  };

  componentDidMount() {
    backend
      .get("/api/status")
      .then((res) => this.setState({ status: res.json }));
  }

  componentWillUnmount() {}

  render() {
    return (
      <div className="status-bar">
        <Link to={"/capture" + this.props.location.search}>
          <div className="live-capture-button">
            {this.state.status.live_capture === "local" ||
            this.state.status.live_capture === "remote" ? (
              <>
                <div className="ringring"></div>
                <div className="circle"></div>
              </>
            ) : (
              <div className="triangle"></div>
            )}
            <span className="label">live capture</span>
          </div>
        </Link>

        <div className="statistics">
          <div className="record">
            <LinkPopover text="pps" content="packets per second" />: 333
          </div>
          <div className="record">
            <LinkPopover text="cps" content="connections per second" />: 421
          </div>
          <div className="record">
            <LinkPopover
              text="connections"
              content="total number of connections"
            />
            : 231312
          </div>
          <div className="record">
            <LinkPopover text="pcaps" content="number of pcaps analyzed" />: {this.state.status.pcaps_analyzed}
          </div>
        </div>

        <div className="separator"></div>

        <div className="quick-settings">
          <CheckField name="auto_refresh" rounded={false} checked={true} />
        </div>

        <div className="version">
          <a
            href={`https://github.com/eciavatta/caronte/releases/tag/${this.state.status.version}`}
          >
            v{this.state.status.version}
          </a>
        </div>
      </div>
    );
  }
}

export default withRouter(StatusBar);
