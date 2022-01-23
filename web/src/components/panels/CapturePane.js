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

import _ from 'lodash';
import React, {Component} from 'react';
import backend from '../../backend';
import dispatcher from '../../dispatcher';
import {createCurlCommand, validatePort} from '../../utils';
import ButtonField from '../fields/ButtonField';
import ChoiceField from '../fields/ChoiceField';
import NumericField from '../fields/extensions/NumericField';
import InputField from '../fields/InputField';
import TagField from '../fields/TagField';
import TextField from '../fields/TextField';
import LinkPopover from '../objects/LinkPopover';
import './CapturePane.scss';
import './common.scss';

class CapturePane extends Component {
  state = {
    localInterfaces: [],
    localSettings: {
      interface: '',
      included_services: [],
      excluded_services: [],
    },
    remoteInterfaces: [],
    remoteSettings: {
      ssh_config: {
        host: '',
        port: 22,
        user: 'root',
        password: '',
        private_key: '',
        passphrase: '',
        server_public_key: '',
      },
      capture_options: {
        interface: '',
        included_services: [],
        excluded_services: [],
      },
    },
    services: [],
    status: null,
  };

  componentDidMount() {
    this.checkStatus();
    this.loadLocalInterfaces();
    this.loadServices();
    dispatcher.register('notifications', this.handleNotifications);
    document.title = 'caronte:~/capture$';
  }

  componentWillUnmount() {
    dispatcher.unregister(this.handleNotifications);
  }

  handleNotifications = (payload) => {
    if (payload.event.startsWith('services')) {
      this.loadServices();
    } else if (payload.event.startsWith('capture')) {
      this.resetState();
      if (payload.event === 'capture.local') {
        this.setState({status: 'local'});
      } else if (payload.event === 'capture.remote') {
        this.setState({status: 'remote'});
      } else if (payload.event === 'capture.stop') {
        this.setState({status: ''});
      }
    }
  };

  resetState = () => {
    this.setState({
      startLocalCaptureStatusCode: null,
      startLocalCaptureResponse: null,
      startRemoteCaptureStatusCode: null,
      startRemoteCaptureResponse: null,
      stopCaptureStatusCode: null,
      stopCaptureResponse: null,
      testSSHSettingsError: null,
      captureIntervalStatusCode: null,
      captureIntervalResponse: null,
    });
  };

  checkStatus = () => {
    backend.get('/api/status').then((res) => this.setState({status: res.json.live_capture}));
  };

  loadServices = () => {
    backend.get('/api/services').then((res) =>
      this.setState({
        services: Object.entries(res.json).map(([port, service]) => {
          return {id: port, name: service.name};
        }),
      })
    );
  };

  loadLocalInterfaces = () => {
    backend
      .post('/api/capture/local/interfaces')
      .then((res) => this.setState({localInterfaces: res.json}))
      .catch((res) => this.setState({localInterfaceError: res.json.error}));
  };

  validateSSHSettings = () => {
    for (const field of ['host', 'port', 'user']) {
      if (!this.state.remoteSettings.ssh_config[field]) {
        this.setState({testSSHSettingsError: `${field} required`});
        return false;
      }
    }

    return true;
  };

  loadRemoteInterfaces = () => {
    backend
      .post('/api/capture/remote/interfaces', this.state.remoteSettings.ssh_config)
      .then((res) => this.setState({remoteInterfaces: res.json}))
      .catch((res) => this.setState({testSSHSettingsError: res.json.error}));
  };

  startLocalCapture = () => {
    if (!this.state.localSettings.interface) {
      this.setState({localInterfaceError: 'required'});
      return;
    }

    this.setState({localInterfaceError: null});

    const handle = (res) =>
      this.setState({
        startLocalCaptureStatusCode: res.statusCode,
        startLocalCaptureResponse: JSON.stringify(res.json),
      });

    backend.put('/api/capture/local', this.state.localSettings).then(handle).catch(handle);
  };

  testSSHSettings = () => {
    if (!this.validateSSHSettings()) {
      return;
    }

    this.setState({testSSHSettingsError: null});
    this.loadRemoteInterfaces();
  };

  startRemoteCapture = () => {
    if (!this.state.remoteSettings.capture_options.interface) {
      this.setState({remoteInterfaceError: 'required'});
      return;
    }

    this.setState({remoteInterfaceError: null});

    const handle = (res) =>
      this.setState({
        startRemoteCaptureStatusCode: res.statusCode,
        startRemoteCaptureResponse: JSON.stringify(res.json),
      });

    backend.put('/api/capture/remote', this.state.remoteSettings).then(handle).catch(handle);
  };

  stopCapture = () => {
    const handle = (res) =>
      this.setState({
        stopCaptureStatusCode: res.statusCode,
        stopCaptureResponse: JSON.stringify(res.json),
      });

    backend.delete('/api/capture').then(handle).catch(handle);
  };

  setCaptureInterval = () => {
    if (!this.state.rotationInterval) {
      return this.setState({rotationIntervalError: 'required'});
    }

    this.setState({rotationIntervalError: null});

    const handle = (res) =>
      this.setState({
        captureIntervalStatusCode: res.statusCode,
        captureIntervalResponse: JSON.stringify(res.json),
      });

    backend.put('/api/capture/interval', {rotation_interval: this.state.rotationInterval}).then(handle).catch(handle);
  };

  updateSettings = (type, callback) => {
    const settings = {...this.state[`${type}Settings`]};
    callback(settings);

    if (type === 'local') {
      this.setState({localSettings: settings});
    } else if (type === 'remote') {
      this.setState({remoteSettings: settings});
    }
  };

  render() {
    return (
      <div className="pane-container live-capture-pane">
        {this.state.status !== null && (this.state.status ? this.stopCaptureRender() : this.startCaptureRender())}
      </div>
    );
  }

  startCaptureRender = () => {
    const createPortFilter = (localRemote, includedExcluded) => {
      const srv = `${includedExcluded}_services`;
      return (
        <TagField
          tags={(localRemote === 'local' ? this.state.localSettings[srv] : this.state.remoteSettings.capture_options[srv]).map((port) => {
            const name = this.state.services.find((s) => s.id === port)?.name;
            return {name: name || port, id: port};
          })}
          name={srv}
          allowNew
          inline
          onChange={(tags) => {
            this.updateSettings(localRemote, (settings) => {
              const captureOptions = localRemote === 'local' ? settings : settings.capture_options;
              captureOptions[srv] = _.uniq(tags.map((t) => parseInt(t.id || t.name)).filter((t) => validatePort(t)));
            });
          }}
          suggestions={this.state.services}
        />
      );
    };

    const lSettings = this.state.localSettings;
    const rSettings = this.state.remoteSettings;

    return (
      <div className="double-pane-container">
        <div className="pane-section">
          <div className="section-header">
            <span className="api-request">PUT /api/capture/local</span>
            <span className="api-response">
              <LinkPopover text={this.state.startLocalCaptureStatusCode} content={this.state.startLocalCaptureResponse} placement="left" />
            </span>
          </div>

          <div className="section-content">
            <ChoiceField
              name="interface"
              keys={this.state.localInterfaces}
              values={this.state.localInterfaces}
              value={lSettings.interface}
              onChange={(key) => this.updateSettings('local', (s) => (s.interface = key))}
              inline
              error={this.state.localInterfaceError}
            />
            {createPortFilter('local', 'included')}
            {createPortFilter('local', 'excluded')}

            <div className="section-footer">
              <ButtonField variant="blue" bordered onClick={this.startLocalCapture} name="start" />
            </div>

            <div className="helper">
              <span>local</span>
            </div>
            <TextField value={createCurlCommand('/capture/local', 'PUT', lSettings)} rows={4} readonly small />

            <div className="helper">
              <span>remote</span>
            </div>
            <TextField value={createCurlCommand('/capture/remote', 'PUT', rSettings)} rows={4} readonly small />
          </div>

          <div className="section-header">
            <span className="api-request">PUT /api/capture/interval</span>
            <span className="api-response">
              <LinkPopover text={this.state.captureIntervalStatusCode} content={this.state.captureIntervalResponse} placement="left" />
            </span>
          </div>

          <div className="section-content">
            <div className="rotation-interval">
              <NumericField
                name="rotation_interval"
                inline
                value={this.state.rotationInterval}
                onChange={(v) => this.setState({rotationInterval: v})}
                min={0}
                max={65565}
                error={this.state.rotationIntervalError}
              />
              <span className="mesurement-unit">seconds</span>
            </div>
          </div>

          <div className="section-footer">
            <ButtonField variant="green" bordered onClick={this.setCaptureInterval} name="save" />
          </div>
        </div>

        <div className="pane-section">
          <div className="section-header">
            <span className="api-request">PUT /api/capture/remote</span>
            <span className="api-response">
              <LinkPopover text={this.state.startRemoteCaptureStatusCode} content={this.state.startRemoteCaptureResponse} placement="left" />
            </span>
          </div>

          <div className="section-content">
            <div className="helper">
              <span>connection settings</span>
            </div>
            <div style={{display: 'flex'}}>
              <InputField
                name="host"
                inline
                style={{flex: '10'}}
                value={rSettings.ssh_config.host}
                onChange={(v) => this.updateSettings('remote', (s) => (s.ssh_config.host = v))}
              />
              <NumericField
                name="port"
                inline
                value={rSettings.ssh_config.port}
                onChange={(v) => this.updateSettings('remote', (s) => (s.ssh_config.port = v))}
                min={0}
                max={65565}
              />
            </div>

            <div style={{display: 'flex'}}>
              <InputField
                name="user"
                inline
                value={rSettings.ssh_config.user}
                onChange={(v) => this.updateSettings('remote', (s) => (s.ssh_config.user = v))}
              />
              <InputField
                name="password"
                inline
                type="password"
                value={rSettings.ssh_config.password}
                onChange={(v) => this.updateSettings('remote', (s) => (s.ssh_config.password = v))}
              />
            </div>

            <TextField
              name="private_key"
              rows={2}
              small
              value={rSettings.ssh_config.private_key}
              onChange={(v) => this.updateSettings('remote', (s) => (s.ssh_config.private_key = v))}
            />
            <InputField
              name="passphrase"
              inline
              type="password"
              value={rSettings.ssh_config.passphrase}
              onChange={(v) => this.updateSettings('remote', (s) => (s.ssh_config.passphrase = v))}
            />

            <TextField
              name="server_public_key"
              rows={2}
              small
              value={rSettings.ssh_config.server_public_key}
              onChange={(v) => this.updateSettings('remote', (s) => (s.ssh_config.server_public_key = v))}
            />

            <div className="section-footer">
              <div className="test-ssh-message">
                {this.state.testSSHSettingsError ? (
                  <span className="error-message">error: {this.state.testSSHSettingsError}</span>
                ) : this.state.remoteInterfaces.length > 0 ? (
                  <span className="success-message">connected.</span>
                ) : null}
              </div>
              <ButtonField variant="green" bordered onClick={this.testSSHSettings} name="test" />
            </div>

            <div className="helper">
              <span>capture settings</span>
            </div>

            <ChoiceField
              name="interface"
              keys={this.state.remoteInterfaces}
              values={this.state.remoteInterfaces}
              value={this.state.remoteSettings.capture_options.interface}
              error={this.state.remoteInterfaceError}
              onChange={(key) => this.updateSettings('remote', (s) => (s.capture_options.interface = key))}
              inline
            />
            {createPortFilter('remote', 'included')}
            {createPortFilter('remote', 'excluded')}

            <div className="section-footer">
              <ButtonField variant="blue" bordered onClick={this.startRemoteCapture} name="start" />
            </div>
          </div>
        </div>
      </div>
    );
  };

  stopCaptureRender() {
    return (
      <div className="pane-section service-edit">
        <div className="section-header">
          <span className="api-request">DELETE /api/capture</span>
          <span className="api-response">
            <LinkPopover text={this.state.stopCaptureStatusCode} content={this.state.stopCaptureResponse} placement="left" />
          </span>
        </div>

        <div className="section-content">
          A {this.state.status} capture is currently in progress. Do you want to stop it?
          <div className="stop-capture-button">
            <ButtonField variant="red" name="stop" bordered onClick={this.stopCapture} />
          </div>
          <TextField value={createCurlCommand('/capture', 'DELETE')} rows={3} readonly small />
        </div>
      </div>
    );
  }
}

export default CapturePane;
