import React, {Component} from 'react';
import './ConnectionContent.scss';
import {Col, Container, Dropdown, Row} from 'react-bootstrap';
import axios from 'axios';
import {withRouter} from "react-router-dom";
import {Redirect} from "react-router";

class ConnectionContent extends Component {

    constructor(props) {
        super(props);
        this.state = {
            loading: false,
            connectionContent: null,
            format: "default"
        };

        this.validFormats = ["default", "hex", "hexdump", "base32", "base64", "ascii", "binary", "decimal", "octal"];
        this.setFormat = this.setFormat.bind(this);
    }

    componentDidMount() {
        if ('format' in this.props.match.params) {
            this.setFormat(this.props.match.params.format);
        }
    }

    componentDidUpdate(prevProps, prevState, snapshot) {
        if (this.props.connection !== null && (
            this.props.connection !== prevProps.connection || this.state.format !== prevState.format)) {
            this.setState({loading: true});
            axios.get(`/api/streams/${this.props.connection.id}?format=${this.state.format}`).then(res => {
                this.setState({
                    connectionContent: res.data,
                    loading: false
                });
            });
        }
    }

    setFormat(format) {
        if (this.validFormats.includes(format)) {
            this.setState({format: format});
        }
    }

    render() {
        let content = this.state.connectionContent;

        if (content === null) {
            return <div>nope</div>;
        }

        const format = this.state.format !== "default" ? `/${this.state.format}` : "";
        const redirect = <Redirect push to={`/connections/${this.props.connection.id}${format}`}/>;

        let payload = content.map((c, i) =>
            <span key={`content-${i}`} className={c.from_client ? "from-client" : "from-server"} title="cccccc">
                {c.content}
            </span>
        );

        return (
            <div className="connection-content">
                <div className="connection-content-options">
                    <Container>
                        <Row>
                            <Col md={2}>ciao</Col>
                        </Row>
                    </Container>


                    <Dropdown onSelect={this.setFormat} >
                        <Dropdown.Toggle size="sm" id="dropdown-basic">
                            format
                        </Dropdown.Toggle>

                        <Dropdown.Menu>
                            <Dropdown.Item eventKey="default" active={this.state.format === "default"}>plain</Dropdown.Item>
                            <Dropdown.Item eventKey="hex" active={this.state.format === "hex"}>hex</Dropdown.Item>
                            <Dropdown.Item eventKey="hexdump" active={this.state.format === "hexdump"}>hexdump</Dropdown.Item>
                            <Dropdown.Item eventKey="base32" active={this.state.format === "base32"}>base32</Dropdown.Item>
                            <Dropdown.Item eventKey="base64" active={this.state.format === "base64"}>base64</Dropdown.Item>
                            <Dropdown.Item eventKey="ascii" active={this.state.format === "ascii"}>ascii</Dropdown.Item>
                            <Dropdown.Item eventKey="binary" active={this.state.format === "binary"}>binary</Dropdown.Item>
                            <Dropdown.Item eventKey="decimal" active={this.state.format === "decimal"}>decimal</Dropdown.Item>
                            <Dropdown.Item eventKey="octal" active={this.state.format === "octal"}>octal</Dropdown.Item>
                        </Dropdown.Menu>
                    </Dropdown>


                </div>

                <pre>{payload}</pre>

                {redirect}


            </div>
        );
    }

}


export default withRouter(ConnectionContent);
