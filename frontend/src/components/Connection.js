import React, {Component} from 'react';
import './Connection.scss';
import {Button, OverlayTrigger, Tooltip} from "react-bootstrap";

class Connection extends Component {
    render() {
        let conn = this.props.data
        let serviceName = "/dev/null"
        let serviceColor = "#0F192E"
        if (conn.service.port !== 0) {
            serviceName = conn.service.name
            serviceColor = conn.service.color
        }
        let startedAt = new Date(conn.started_at)
        let closedAt = new Date(conn.closed_at)
        let duration = ((closedAt - startedAt) / 1000).toFixed(3)
        let timeInfo = `Started at ${startedAt}\nClosed at ${closedAt}\nProcessed at ${new Date(conn.processed_at)}`

        let classes = "connection"
        if (this.props.selected) {
            classes += " connection-selected"
        }
        if (conn.marked){
            classes += " connection-marked"
        }

        return (
            <tr className={classes}>
                <td>
                    <span className="connection-service">
                        <Button size="sm" style={{
                            "backgroundColor": serviceColor
                        }}>{serviceName}</Button>
                    </span>
                </td>
                <td className="clickable" onClick={() => this.props.onSelected()}>{conn.ip_src}</td>
                <td className="clickable" onClick={() => this.props.onSelected()}>{conn.port_src}</td>
                <td className="clickable" onClick={() => this.props.onSelected()}>{conn.ip_dst}</td>
                <td className="clickable" onClick={() => this.props.onSelected()}>{conn.port_dst}</td>
                <td className="clickable" onClick={() => this.props.onSelected()}>
                    {/*<OverlayTrigger placement="top" overlay={<Tooltip id={`tooltip-${conn.id}`}>{timeInfo}</Tooltip>}>*/}
                        <span className="test-tooltip">{duration}s</span>
                    {/*</OverlayTrigger>*/}
                </td>
                <td className="clickable" onClick={() => this.props.onSelected()}>{conn.client_bytes}</td>
                <td className="clickable" onClick={() => this.props.onSelected()}>{conn.server_bytes}</td>
                <td>
                    <span className="connection-icon connection-hide">%</span>
                    <span className="connection-icon connection-mark">!!</span>
                    <span className="connection-icon connection-comment">@</span>
                    <span className="connection-icon connection-link">#</span>
                </td>

            </tr>
        );
    }

}


export default Connection;
