import React, {Component} from 'react';
import './Services.scss';
import {Button, ButtonGroup, Col, Container, Form, FormControl, InputGroup, Modal, Row, Table} from "react-bootstrap";
import axios from "axios";

class Rules extends Component {

    constructor(props) {
        super(props);

        this.state = {
            rules: []
        };
    }

    componentDidMount() {
        this.loadRules();
    }

    loadRules() {
        axios.get("/api/rules").then(res => this.setState({rules: res.data}));
    }

    render() {
        let rulesRows = this.state.rules.map(rule =>
            <tr key={rule.id}>
                <td><Button variant="btn-edit" size="sm"
                            style={{"backgroundColor": rule.color}}>edit</Button></td>
                <td>{rule.name}</td>
            </tr>
        );


        return (
            <Modal
                {...this.props}
                show="true"
                size="lg"
                aria-labelledby="rules-dialog"
                centered
            >
                <Modal.Header>
                    <Modal.Title id="rules-dialog">
                        ~/rules
                    </Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <Container>
                        <Row>
                            <Col md={7}>
                                <Table borderless size="sm" className="rules-list">
                                    <thead>
                                    <tr>
                                        <th><Button size="sm" >new</Button></th>
                                        <th>name</th>
                                    </tr>
                                    </thead>
                                    <tbody>
                                    {rulesRows}

                                    </tbody>
                                </Table>
                            </Col>
                            <Col md={5}>
                                <Form>
                                    <Form.Group controlId="servicePort">
                                        <Form.Label>port:</Form.Label>
                                        <Form.Control type="text"  />
                                    </Form.Group>

                                    <Form.Group controlId="serviceName">
                                        <Form.Label>name:</Form.Label>
                                        <Form.Control type="text"  />
                                    </Form.Group>

                                    <Form.Group controlId="serviceColor">
                                        <Form.Label>color:</Form.Label>
                                        <ButtonGroup aria-label="Basic example">

                                        </ButtonGroup>
                                        <ButtonGroup aria-label="Basic example">

                                        </ButtonGroup>
                                    </Form.Group>

                                    <Form.Group controlId="serviceNotes">
                                        <Form.Label>notes:</Form.Label>
                                        <Form.Control as="textarea" rows={3} />
                                    </Form.Group>
                                </Form>


                            </Col>

                        </Row>

                        <Row>
                            <Col md={12}>
                                <InputGroup>
                                    <FormControl as="textarea" rows={4} className="curl-output" readOnly={true}
                                                 />
                                </InputGroup>

                            </Col>
                        </Row>


                    </Container>
                </Modal.Body>
                <Modal.Footer className="dialog-footer">
                    <Button variant="red" onClick={this.props.onHide}>close</Button>
                </Modal.Footer>
            </Modal>
        );
    }
}

export default Rules;
