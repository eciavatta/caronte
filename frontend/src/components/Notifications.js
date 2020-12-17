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
import dispatcher from "../dispatcher";
import {randomClassName} from "../utils";
import "./Notifications.scss";

const _ = require("lodash");
const classNames = require("classnames");

class Notifications extends Component {

    state = {
        notifications: [],
        closedNotifications: [],
    };

    componentDidMount() {
        dispatcher.register("notifications", this.handleNotifications);
    }

    componentWillUnmount() {
        dispatcher.unregister(this.handleNotifications);
    }

    handleNotifications = (n) => this.notificationHandler(n);

    notificationHandler = (n) => {
        switch (n.event) {
            case "connected":
                n.title = "connected";
                n.description = `number of active clients: ${n.message["connected_clients"]}`;
                return this.pushNotification(n);
            case "services.edit":
                n.title = "services updated";
                n.description = `updated "${n.message["name"]}" on port ${n.message["port"]}`;
                n.variant = "blue";
                return this.pushNotification(n);
            case "rules.new":
                n.title = "rules updated";
                n.description = `new rule added: ${n.message["name"]}`;
                n.variant = "green";
                return this.pushNotification(n);
            case "rules.edit":
                n.title = "rules updated";
                n.description = `existing rule updated: ${n.message["name"]}`;
                n.variant = "blue";
                return this.pushNotification(n);
            case "pcap.completed":
                n.title = "new pcap analyzed";
                n.description = `${n.message["processed_packets"]} packets processed`;
                n.variant = "blue";
                return this.pushNotification(n);
            case "timeline.range.large":
                n.title = "timeline cropped";
                n.description = `the maximum range is 24h`;
                n.variant = "red";
                return this.pushNotification(n);
            default:
                return null;
        }
    };

    pushNotification = (notification) => {
        const notifications = this.state.notifications;
        notification.id = randomClassName();
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
    };

    render() {
        return (
            <div className="notifications">
                <div className="notifications-list">
                    {
                        this.state.closedNotifications.concat(this.state.notifications).map((n) => {
                            const notificationClassnames = {
                                "notification": true,
                                "notification-closed": n.closed,
                                "notification-open": n.open
                            };
                            if (n.variant) {
                                notificationClassnames[`notification-${n.variant}`] = true;
                            }
                            return <div key={n.id} className={classNames(notificationClassnames)} onClick={n.onClick}>
                                <h3 className="notification-title">{n.title}</h3>
                                <pre className="notification-description">{n.description}</pre>
                            </div>;
                        })
                    }
                </div>
            </div>
        );
    }
}

export default Notifications;
