import React, {Component} from 'react';
import './Services.scss';
import {Button, ButtonGroup, Col, Container, Form, FormControl, InputGroup, Modal, Row, Table} from "react-bootstrap";
import axios from 'axios'
import {createCurlCommand} from '../utils';

class Services extends Component {

    constructor(props) {
        super(props);
        this.alphabet = 'abcdefghijklmnopqrstuvwxyz';
        this.colors = ["#E53935", "#D81B60", "#8E24AA", "#5E35B1", "#3949AB", "#1E88E5", "#039BE5", "#00ACC1",
            "#00897B", "#43A047", "#7CB342", "#9E9D24", "#F9A825", "#FB8C00", "#F4511E", "#6D4C41"];

        this.state = {
            services: {},
            port: 0,
            portValid: false,
            name: "",
            nameValid: false,
            color: this.colors[0],
            colorValid: false,
            notes: ""
        };

        this.portChanged = this.portChanged.bind(this);
        this.nameChanged = this.nameChanged.bind(this);
        this.notesChanged = this.notesChanged.bind(this);
        this.newService = this.newService.bind(this);
        this.editService = this.editService.bind(this);
        this.saveService = this.saveService.bind(this);
        this.loadServices = this.loadServices.bind(this);
    }

    componentDidMount() {
        this.loadServices();
    }

    portChanged(event) {
        let value = event.target.value.replace(/[^\d]/gi, '');
        let port = 0;
        if (value !== "") {
            port = parseInt(value);
        }
        this.setState({port: port});
    }

    nameChanged(event) {
        let value = event.target.value.replace(/[\s]/gi, '_').replace(/[^\w]/gi, '').toLowerCase();
        this.setState({name: value});
    }

    notesChanged(event) {
        this.setState({notes: event.target.value});
    }

    newService() {
        this.setState({name: "", port: 0, notes: ""});
    }

    editService(service) {
        this.setState({name: service.name, port: service.port, color: service.color, notes: service.notes});
    }

    saveService() {
        if (this.state.portValid && this.state.nameValid) {
            axios.put("/api/services", {
                name: this.state.name,
                port: this.state.port,
                color: this.state.color,
                notes: this.state.notes
            });

            this.newService();
            this.loadServices();
        }
    }

    loadServices() {
        axios.get("/api/services").then(res => this.setState({services: res.data}));
    }

    componentDidUpdate(prevProps, prevState, snapshot) {
        if (this.state.name != null && prevState.name !== this.state.name) {
            this.setState({
                nameValid: this.state.name.length >= 3
            });
        }
        if (prevState.port !== this.state.port) {
            this.setState({
                portValid: this.state.port > 0 && this.state.port <= 65565
            });
        }
    }

    render() {
        let output = "";
        if (!this.state.portValid) {
            output += "assert(1 <= port <= 65565)\n";
        }
        if (!this.state.nameValid) {
            output += "assert(len(name) >= 3)\n";
        }
        if (output === "") {
            output = createCurlCommand("/services", {
                "port": this.state.port,
                "name": this.state.name,
                "color": this.state.color,
                "notes": this.state.notes
            });
        }
        let rows = Object.values(this.state.services).map(s =>
            <tr>
                <td><Button variant="btn-edit" size="sm"
                            onClick={() => this.editService(s)} style={{ "backgroundColor": s.color }}>edit</Button></td>
                <td>{s.port}</td>
                <td>{s.name}</td>
            </tr>
        );

        let colorButtons = this.colors.map((color, i) =>
            <Button size="sm" className="btn-color" key={"button" + this.alphabet[i]}
                    style={{"backgroundColor": color, "borderColor": this.state.color === color ? "#fff" : color}}
                    onClick={() => this.setState({color: color})}>{this.alphabet[i]}</Button>);

        return (
            <Modal
                {...this.props}
                show="true"
                size="lg"
                aria-labelledby="services-dialog"
                centered
            >
                <Modal.Header>
                    <Modal.Title id="services-dialog">
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
                                        <th><Button size="sm" onClick={this.newService}>new</Button></th>
                                        <th>port</th>
                                        <th>name</th>
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
                                    </Form.Group>

                                    <Form.Group controlId="serviceName">
                                        <Form.Label>name:</Form.Label>
                                        <Form.Control type="text" onChange={this.nameChanged} value={this.state.name}/>
                                    </Form.Group>

                                    <Form.Group controlId="serviceColor">
                                        <Form.Label>color:</Form.Label>
                                        <ButtonGroup aria-label="Basic example">
                                            {colorButtons.slice(0,8)}
                                        </ButtonGroup>
                                        <ButtonGroup aria-label="Basic example">
                                            {colorButtons.slice(8,18)}
                                        </ButtonGroup>
                                    </Form.Group>

                                    <Form.Group controlId="serviceNotes">
                                        <Form.Label>notes:</Form.Label>
                                        <Form.Control as="textarea" rows={3} onChange={this.notesChanged} value={this.state.notes} />
                                    </Form.Group>
                                </Form>


                            </Col>

                        </Row>

                        <Row>
                            <Col md={12}>
                                <InputGroup>
                                    <FormControl as="textarea" rows={4} className="curl-output" readOnly={true} value={output}/>
                                </InputGroup>

                            </Col>
                        </Row>
                    </Container>
                </Modal.Body>
                <Modal.Footer className="dialog-footer">
                    <Button variant="green" onClick={this.saveService}>save</Button>
                    <Button variant="red" onClick={this.props.onHide}>close</Button>
                </Modal.Footer>
            </Modal>
        );
    }
}

export default Services;
