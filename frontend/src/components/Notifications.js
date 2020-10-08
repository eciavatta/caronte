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
import './Notifications.scss';
import dispatcher from "../dispatcher";

const _ = require('lodash');
const classNames = require('classnames');

class Notifications extends Component {

    state = {
        notifications: [],
        closedNotifications: [],
    };

    componentDidMount() {
        dispatcher.register("notifications", notification => {
            const notifications = this.state.notifications;
            notifications.push(notification);
            this.setState({notifications});
            setTimeout(() => {
                const notifications = this.state.notifications;
                notification.open = true;
                this.setState({notifications});
            }, 100);

            const hideHandle = setTimeout(() => {
                const notifications = _.without(this.state.notifications, notification);
                const closedNotifications = this.state.closedNotifications.concat([notification]);
                notification.closed = true;
                this.setState({notifications, closedNotifications});
            }, 5000);

            const removeHandle = setTimeout(() => {
                const closedNotifications = _.without(this.state.closedNotifications, notification);
                this.setState({closedNotifications});
            }, 6000);

            notification.onClick = () => {
                clearTimeout(hideHandle);
                clearTimeout(removeHandle);
                const notifications = _.without(this.state.notifications, notification);
                this.setState({notifications});
            };
        });
    }

    render() {
        return (
            <div className="notifications">
                <div className="notifications-list">
                    {
                        this.state.closedNotifications.concat(this.state.notifications).map(n =>
                            <div className={classNames("notification", {"notification-closed": n.closed},
                                {"notification-open": n.open})} onClick={n.onClick}>
                                <h3 className="notification-title">{n.event}</h3>
                                <span className="notification-description">{JSON.stringify(n.message)}</span>
                            </div>
                        )
                    }
                </div>
            </div>
        );
    }
}

export default Notifications;
