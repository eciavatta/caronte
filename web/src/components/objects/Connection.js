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
import backend from '../../backend';
import dispatcher from '../../dispatcher';
import {calculateSimilarityScores, dateTimeToTime, durationBetween, formatSize, generateSimilarityProps} from '../../utils';
import CommentDialog from '../dialogs/CommentDialog';
import ButtonField from '../fields/ButtonField';
import TextField from '../fields/TextField';
import './Connection.scss';
import CopyLinkPopover from './CopyLinkPopover';
import LinkPopover from './LinkPopover';

class Connection extends Component {
  state = {
    update: false,
  };

  static get propTypes() {
    return {
      data: PropTypes.object,
      onCommented: PropTypes.func,
      onMarked: PropTypes.func,
      onSelected: PropTypes.func,
      selected: PropTypes.bool,
      selectedSimilarityProps: PropTypes.object,
      services: PropTypes.array,
    };
  }

  handleAction = (name, comment) => {
    if (name === 'mark') {
      const marked = this.props.data.marked;
      backend.post(`/api/connections/${this.props.data.id}/${marked ? 'unmark' : 'mark'}`).then(() => {
        this.props.onMarked(!marked);
        this.setState({update: true});
      });
    } else if (name === 'comment') {
      this.props.onCommented(comment);
      this.setState({showCommentDialog: false});
    }
  };

  render() {
    let conn = this.props.data;
    let serviceName = '/dev/null';
    let serviceColor = '#0f192e';
    if (this.props.services[conn['port_dst']]) {
      const service = this.props.services[conn['port_dst']];
      serviceName = service.name;
      serviceColor = service.color;
    }
    let startedAt = new Date(conn['started_at']);
    let closedAt = new Date(conn['closed_at']);
    let processedAt = new Date(conn['processed_at']);
    let timeInfo = (
      <div>
        <span>Started at {startedAt.toLocaleDateString() + ' ' + startedAt.toLocaleTimeString()}</span>
        <br />
        <span>Processed at {processedAt.toLocaleDateString() + ' ' + processedAt.toLocaleTimeString()}</span>
        <br />
        <span>Closed at {closedAt.toLocaleDateString() + ' ' + closedAt.toLocaleTimeString()}</span>
      </div>
    );

    const commentPopoverContent = (
      <div style={{width: '250px'}}>
        <span>
          Click to <strong>{conn.comment ? 'edit' : 'add'}</strong> comment
        </span>
        {conn.comment && <TextField rows={3} value={conn.comment} readonly />}
      </div>
    );

    const distanceScores = this.props.selectedSimilarityProps && calculateSimilarityScores(this.props.selectedSimilarityProps, generateSimilarityProps(conn));

    return (
      <tr className={classNames('connection', {'connection-selected': this.props.selected}, {'has-matched-rules': conn.matched_rules.length > 0})}>
        <td>
          <span className="connection-service">
            <ButtonField
              small
              fullSpan
              color={serviceColor}
              name={serviceName}
              onClick={() => dispatcher.dispatch('connections_filters', {service_port: conn['port_dst'].toString()})}
            />
          </span>
        </td>
        <td className="clickable" onClick={this.props.onSelected}>
          {conn['ip_src']}
        </td>
        <td className="clickable" onClick={this.props.onSelected}>
          {conn['port_src']}
        </td>
        <td className="clickable" onClick={this.props.onSelected}>
          {conn['ip_dst']}
        </td>
        <td className="clickable" onClick={this.props.onSelected}>
          {conn['port_dst']}
        </td>
        <td className="clickable" onClick={this.props.onSelected}>
          <LinkPopover text={dateTimeToTime(conn['started_at'])} content={timeInfo} placement="right" />
        </td>
        <td className="clickable" onClick={this.props.onSelected}>
          {durationBetween(startedAt, closedAt)}
        </td>
        <td className="clickable" onClick={this.props.onSelected}>
          {formatSize(conn['client_bytes'])}
        </td>
        <td className="clickable" onClick={this.props.onSelected}>
          {formatSize(conn['server_bytes'])}
        </td>
        <td>
          {(distanceScores && (
            <LinkPopover
              text={
                <span
                  onClick={() =>
                    dispatcher.dispatch('connections_filters', {similar_to_id: null, client_similar_to_id: conn['id'], server_similar_to_id: null})
                  }
                >
                  {distanceScores['clientSimilarityScore']['score']}
                </span>
              }
              content={
                <>
                  Click to find <strong>similar</strong> connections with this client connection stream
                  <br />
                  <br />
                  {JSON.stringify(distanceScores['clientSimilarityScore'])}
                </>
              }
              placement="right"
            />
          )) ||
            '-'}
        </td>
        <td>
          {(distanceScores && (
            <LinkPopover
              text={
                <span
                  onClick={() =>
                    dispatcher.dispatch('connections_filters', {similar_to_id: null, client_similar_to_id: null, server_similar_to_id: conn['id']})
                  }
                >
                  {distanceScores['serverSimilarityScore']['score']}
                </span>
              }
              content={
                <>
                  Click to find <strong>similar</strong> connections with this server connection stream
                  <br />
                  <br />
                  {JSON.stringify(distanceScores['serverSimilarityScore'])}
                </>
              }
              placement="right"
            />
          )) ||
            '-'}
        </td>
        <td className="connection-actions">
          <LinkPopover
            text={
              <span className={classNames('connection-icon', {'icon-enabled': conn.marked})} onClick={() => this.handleAction('mark')}>
                !!
              </span>
            }
            content={<span>Mark this connection</span>}
            placement="right"
          />
          <LinkPopover
            text={
              <span className={classNames('connection-icon', {'icon-enabled': conn.comment})} onClick={() => this.setState({showCommentDialog: true})}>
                @
              </span>
            }
            content={commentPopoverContent}
            placement="right"
          />
          <CopyLinkPopover text="#" value={conn.id} textClassName={classNames('connection-icon', {'icon-enabled': conn.hidden})} />
          {this.state.showCommentDialog && (
            <CommentDialog onSave={(comment) => this.handleAction('comment', comment)} initialComment={conn.comment} connectionId={conn.id} />
          )}
          <LinkPopover
            text={
              <span
                className={classNames('connection-icon')}
                onClick={() =>
                  dispatcher.dispatch('connections_filters', {
                    similar_to_id: conn['id'],
                    client_similar_to_id: null,
                    server_similar_to_id: null,
                  })
                }
              >
                $
              </span>
            }
            content={
              <>
                Click to find <strong>similar</strong> connections
              </>
            }
          />
        </td>
      </tr>
    );
  }
}

export default Connection;
