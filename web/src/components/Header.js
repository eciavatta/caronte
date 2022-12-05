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

import classNames from 'classnames';
import PropTypes from 'prop-types';
import React, {Component} from 'react';
import {Link} from 'react-router-dom';
import Typed from 'typed.js';
import {cleanNumber, validatePort, withRouter} from '../utils';
import ButtonField from './fields/ButtonField';
import AdvancedFilters from './filters/AdvancedFilters';
import BooleanConnectionsFilter from './filters/BooleanConnectionsFilter';
import ExitSearchFilter from './filters/ExitSearchFilter';
import ExitSimilarityFilter from './filters/ExitSimilarityFilter';
import RulesConnectionsFilter from './filters/RulesConnectionsFilter';
import StringConnectionsFilter from './filters/StringConnectionsFilter';
import './Header.scss';

class Header extends Component {
  static get propTypes() {
    return {
      configured: PropTypes.bool,
      location: PropTypes.object,
      onOpenFilters: PropTypes.func,
    };
  }

  componentDidMount() {
    const options = {
      strings: ['caronte$ '],
      typeSpeed: 50,
      cursorChar: '❚',
    };
    this.typed = new Typed(this.el, options);
  }

  componentWillUnmount() {
    this.typed.destroy();
  }

  render() {
    return (
      <header className={classNames('header', {configured: this.props.configured})}>
        <div className="header-content">
          <h1 className="header-title type-wrap">
            <Link to="/">
              <span
                style={{whiteSpace: 'pre'}}
                ref={(el) => {
                  this.el = el;
                }}
              />
            </Link>
          </h1>

          {this.props.configured && (
            <div className="filters-bar">
              <StringConnectionsFilter
                filterName="service_port"
                defaultFilterValue="all_ports"
                replaceFunc={cleanNumber}
                validateFunc={validatePort}
                key="service_port_filter"
                width={200}
                small
                inline
              />
              <RulesConnectionsFilter />
              <BooleanConnectionsFilter filterName={'marked'} />
              <ExitSearchFilter />
              <ExitSimilarityFilter />
              <AdvancedFilters onClick={this.props.onOpenFilters} />
            </div>
          )}

          {this.props.configured && (
            <div className="header-buttons">
              <Link to={'/searches' + this.props.router.location.search}>
                <ButtonField variant="pink" name="searches" bordered />
              </Link>
              <Link to={'/pcaps' + this.props.router.location.search}>
                <ButtonField variant="purple" name="pcaps" bordered />
              </Link>
              <Link to={'/rules' + this.props.router.location.search}>
                <ButtonField variant="deep-purple" name="rules" bordered />
              </Link>
              <Link to={'/services' + this.props.router.location.search}>
                <ButtonField variant="indigo" name="services" bordered />
              </Link>
              <Link to={'/stats' + this.props.router.location.search}>
                <ButtonField variant="blue" name="stats" bordered />
              </Link>
            </div>
          )}
        </div>
      </header>
    );
  }
}

export default withRouter(Header);
