import React, {Component} from 'react';
import './Connection.scss';
import {FontAwesomeIcon} from "@fortawesome/react-fontawesome";
import {
    faCloudDownloadAlt,
    faCloudUploadAlt,
    faComment,
    faEyeSlash,
    faHourglassHalf,
    faLaptop,
    faLink,
    faServer,
    faThumbtack,
} from '@fortawesome/free-solid-svg-icons'

class Connection extends Component {
    render() {
        let conn = this.props.data
        let serviceName = "assign"
        let serviceColor = "#fff"
        if (conn.service != null) {
            serviceName = conn.service.name
            serviceColor = conn.service.color
        }
        let startedAt = new Date(conn.started_at)
        let closedAt = new Date(conn.closed_at)
        let duration = ((closedAt - startedAt) / 1000).toFixed(3)
        let timeInfo = `Started at ${startedAt}\nClosed at ${closedAt}\nProcessed at ${new Date(conn.processed_at)}`


        return (
            <tr className={conn.marked ? "connection connection-marked" : "connection"}>
                <div className="connection-header">
                    <span className="connection-service">
                        <button className="btn" style={{
                            "backgroundColor": serviceColor
                        }}>{serviceName}</button>
                    </span>
                    <span className="connection-src">
                        <FontAwesomeIcon icon={faLaptop}/>
                        <span className="connection-ip-port">{conn.ip_src}:{conn.port_src}</span>
                    </span>
                    <span className="connection-separator">{"->"}</span>
                    <span className="connection-dst">
                        <FontAwesomeIcon icon={faServer}/>
                        <span className="connection-ip-port">{conn.ip_dst}:{conn.port_dst}</span>
                    </span>

                    <span className="connection-duration" data-toggle="tooltip" data-placement="top" title={timeInfo}>
                        <FontAwesomeIcon icon={faHourglassHalf}/>
                        <span className="connection-seconds">{duration}s</span>
                    </span>
                    <span className="connection-bytes">
                        <FontAwesomeIcon icon={faCloudDownloadAlt}/>
                        <span className="connection-bytes-count">{conn.client_bytes}</span>
                        <FontAwesomeIcon icon={faCloudUploadAlt}/>
                        <span className="connection-bytes-count">{conn.server_bytes}</span>
                    </span>
                    <span className="connection-hide"><FontAwesomeIcon icon={faEyeSlash}/></span>
                    <span className="connection-mark"><FontAwesomeIcon icon={faThumbtack}/></span>
                    <span className="connection-comment"><FontAwesomeIcon icon={faComment}/></span>
                    <span className="connection-link"><FontAwesomeIcon icon={faLink}/></span>
                </div>


            </tr>
        );
    }

}


export default Connection;
