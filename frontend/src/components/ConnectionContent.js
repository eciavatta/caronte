import React, {Component} from 'react';
import './ConnectionContent.scss';
import {Dropdown} from 'react-bootstrap';

class ConnectionContent extends Component {
    render() {
        let content = this.props.connectionPayload;

        if (content === undefined) {
            return <div>nope</div>;
        }

        let payload = content.map(c =>
            <span key={c.id} className={c.from_client ? "from-client" : "from-server"} title="cccccc">
                {c.content}

            </span>
        );

        return (
            <div className="connection-content">
                <div className="connection-content-options">
                    <Dropdown>
                        <Dropdown.Toggle variant="success" id="dropdown-basic">
                            format
                        </Dropdown.Toggle>

                        <Dropdown.Menu>
                            <Dropdown.Item href="#/action-1">plain</Dropdown.Item>
                            <Dropdown.Item href="#/action-2">hex</Dropdown.Item>
                            <Dropdown.Item href="#/action-3">hexdump</Dropdown.Item>
                            <Dropdown.Item href="#/action-3">base32</Dropdown.Item>
                            <Dropdown.Item href="#/action-3">base64</Dropdown.Item>
                            <Dropdown.Item href="#/action-3">ascii</Dropdown.Item>
                        </Dropdown.Menu>
                    </Dropdown>


                </div>

                <pre>{payload}</pre>


            </div>
        );
    }

}


export default ConnectionContent;
