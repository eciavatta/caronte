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
import {Link} from 'react-router-dom';
import {withRouter} from '../utils';
import dispatcher from '../dispatcher';
import backend from '../backend';
import CheckField from './fields/CheckField';
import LinkPopover from './objects/LinkPopover';
import './StatusBar.scss';

class StatusBar extends Component {
  state = {
    status: {},
    connectionsStatistics: {},
    packetsStatistics: {},
  };

  static get propTypes() {
    return {
      location: PropTypes.object,
    };
  }

  componentDidMount() {
    backend.get('/api/status').then((res) => this.setState({status: res.json}));
    dispatcher.register('notifications', this.handleNotifications);
  }

  componentWillUnmount() {
    dispatcher.unregister(this.handleNotifications);
  }

  handleNotifications = (payload) => {
    const status = {...this.state.status};

    if (payload.event.startsWith('capture')) {
      if (payload.event === 'capture.local') {
        status.live_capture = 'local';
      } else if (payload.event === 'capture.remote') {
        status.live_capture = 'remote';
      } else if (payload.event === 'capture.stop') {
        status.live_capture = null;
      }
    } else if (payload.event === 'connections.statistics') {
      this.setState({connectionsStatistics: payload.message});
    } else if (payload.event === 'packets.statistics') {
      this.setState({packetsStatistics: payload.message});
    }

    this.setState({status});
  };

  render() {
    return (
      <div className="status-bar">
        <Link to={'/capture' + this.props.router.location.search}>
          <div className="live-capture-button">
            {this.state.status.live_capture === 'local' || this.state.status.live_capture === 'remote' ? (
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
            <LinkPopover text="ppm" content="packets per minute" />: - {this.state.packetsStatistics.packets_per_minute || '-'}
          </div>
          <div className="record">
            <LinkPopover text="cpm" content="connections per minute" />: {this.state.connectionsStatistics.connections_per_minute || '-'}
          </div>
          <div className="record">
            <LinkPopover text="pending" content="in-memory connections not already closed" />: {this.state.connectionsStatistics.pending_connections || '-'}
          </div>
          <div className="record">
            <LinkPopover text="connections" content="reassembled connections in current session" />:{' '}
            {this.state.connectionsStatistics.completed_connections || '-'}
          </div>
          <div className="record">
            <LinkPopover text="packets" content="total number of processed packets" />: {this.state.packetsStatistics.processed_packets || '-'}
          </div>
          <div className="record">
            <LinkPopover text="invalid" content="total number of invalid packets" />: {this.state.packetsStatistics.invalid_packets || '-'}
          </div>
        </div>

        <div className="separator"></div>

        <div className="quick-settings">
          <CheckField name="auto_refresh" rounded={false} checked={true} />
        </div>

        <div className="version">
          <a href={`https://github.com/eciavatta/caronte/releases/tag/${this.state.status.version}`}>v{this.state.status.version}</a>
        </div>
      </div>
    );
  }
}

export default withRouter(StatusBar);
