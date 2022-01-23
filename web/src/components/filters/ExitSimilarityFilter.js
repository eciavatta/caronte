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

import PropTypes from 'prop-types';
import React, {Component} from 'react';
import {withRouter} from 'react-router-dom';
import dispatcher from '../../dispatcher';
import CheckField from '../fields/CheckField';

class ExitSimilarityFilter extends Component {
  state = {};

  static get propTypes() {
    return {
      location: PropTypes.object,
      width: PropTypes.number,
    };
  }

  componentDidMount() {
    let params = new URLSearchParams(this.props.location.search);
    this.setState({
      similarityId: params.get('similar_to_id'),
      clientSimilarityId: params.get('client_similar_to_id'),
      serverSimilarityId: params.get('server_similar_to_id'),
    });

    this.connectionsFiltersCallback = (payload) => {
      if ('similar_to_id' in payload && this.state.similarityId !== payload['similar_to_id']) {
        this.setState({
          similarityId: payload['similar_to_id'],
        });
      }
      if ('client_similar_to_id' in payload && this.state.clientSimilarityId !== payload['client_similar_to_id']) {
        this.setState({
          clientSimilarityId: payload['client_similar_to_id'],
        });
      }
      if ('server_similar_to_id' in payload && this.state.serverSimilarityId !== payload['server_similar_to_id']) {
        this.setState({
          serverSimilarityId: payload['server_similar_to_id'],
        });
      }
    };
    dispatcher.register('connections_filters', this.connectionsFiltersCallback);
  }

  componentWillUnmount() {
    dispatcher.unregister(this.connectionsFiltersCallback);
  }

  render() {
    return (
      <>
        {(this.state.similarityId || this.state.clientSimilarityId || this.state.serverSimilarityId) && (
          <div className="filter" style={{width: `${this.props.width}px`}}>
            <CheckField
              checked
              name="exit_similarity"
              onChange={() =>
                dispatcher.dispatch('connections_filters', {
                  similar_to_id: null,
                  client_similar_to_id: null,
                  server_similar_to_id: null,
                })
              }
              small
            />
          </div>
        )}
      </>
    );
  }
}

export default withRouter(ExitSimilarityFilter);
