import React, {Component} from 'react';
import './Services.scss';
import {Button, ButtonGroup, Col, Container, Form, FormControl, InputGroup, Modal, Row, Table} from "react-bootstrap";
import axios from 'axios'
import {createCurlCommand} from '../utils';

class Services extends Component {

    constructor(props) {
        super(props);
        this.state = {
            services: {},
            port: "",
            portValid: false
        }

        this.portChanged = this.portChanged.bind(this);
    }

    componentDidMount() {
        axios.get("/api/services").then(res => this.setState({services: res.data}))
    }

    portChanged(event) {
        let value = event.target.value.replace(/[^\d]/gi, '')
        let intValue = parseInt(value)
        this.setState({port: value, portValid: intValue > 0 && intValue <= 65565})


    }


    render() {
        let curl = createCurlCommand("/services", {
            "port": this.state.port,
            "name": "aaaaa",
            "color": "#fff",
            "notes": "aaa"
        })

        let rows = Object.values(this.state.services).map(s =>
            <tr>
                <td><Button size="sm" style={{
                    "backgroundColor": s.color
                }}>edit</Button></td>
                <td>{s.port}</td>
                <td>{s.name}</td>
            </tr>
        )




        return (
            <Modal
                {...this.props}
                show="true"
                size="lg"
                aria-labelledby="contained-modal-title-vcenter"
                centered
            >
                <Modal.Header>
                    <Modal.Title id="contained-modal-title-vcenter">
                        ~/services
                    </Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <Container>
                        <Row>
                            <Col md={7}>
                                <Table borderless size="sm" className="services-list">
                                    <thead>
                                    <tr>
                                        <th><Button size="sm">new</Button></th>
                                        <th>name</th>
                                        <th>port</th>
                                    </tr>
                                    </thead>
                                    <tbody>
                                        {rows}
                                    </tbody>
                                </Table>
                            </Col>
                            <Col md={5}>
                                <Form>
                                    <Form.Group controlId="servicePort">
                                        <Form.Label>port:</Form.Label>
                                        <Form.Control type="text" onChange={this.portChanged} value={this.state.port} />
                                        <Form.Text className="text-muted">
                                            {!this.state.portValid ? "assert(1 <= port <= 65565)" : ""}
                                        </Form.Text>
                                    </Form.Group>

                                    <Form.Group controlId="serviceName">
                                        <Form.Label>name:</Form.Label>
                                        <Form.Control type="text" required min={3} max={16} />
                                        <Form.Text className="text-muted">
                                            {"assert(len(name) >= 3)"}
                                        </Form.Text>
                                    </Form.Group>

                                    <Form.Group controlId="serviceColor">
                                        <Form.Label>color:</Form.Label>
                                        <ButtonGroup aria-label="Basic example">
                                            <Button variant="secondary">Left</Button>
                                            <Button variant="secondary">Middle</Button>
                                            <Button variant="secondary">Right</Button>
                                        </ButtonGroup>
                                    </Form.Group>

                                    <Form.Group controlId="serviceNotes">
                                        <Form.Label>notes:</Form.Label>
                                        <Form.Control type="textarea" />
                                    </Form.Group>
                                </Form>


                            </Col>

                        </Row>

                        <Row>
                            <Col md={12}>
                                <InputGroup>
                                    <FormControl as="textarea" rows={9} className="curl-output" readOnly={true}>{curl}
                                    </FormControl>
                                </InputGroup>

                            </Col>
                        </Row>
                    </Container>
                </Modal.Body>
                <Modal.Footer className="dialog-footer">
                    <Button variant="green" onClick={this.props.onHide}>save</Button>
                    <Button variant="red" onClick={this.props.onHide}>close</Button>
                </Modal.Footer>
            </Modal>
        );
    }
}

export default Services;
