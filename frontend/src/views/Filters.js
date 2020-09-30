import React, {Component} from 'react';
import {Col, Container, Modal, Row, Table} from "react-bootstrap";
import {filtersDefinitions, filtersNames} from "../components/filters/FiltersDefinitions";
import ButtonField from "../components/fields/ButtonField";

class Filters extends Component {

    constructor(props) {
        super(props);
        let newState = {};
        filtersNames.forEach(elem => newState[`${elem}_active`] = false);
        this.state = newState;
    }

    componentDidMount() {
        let newState = {};
        filtersNames.forEach(elem => newState[`${elem}_active`] = localStorage.getItem(`filters.${elem}`) === "true");
        this.setState(newState);
    }

    checkboxChangesHandler(filterName, event) {
        this.setState({[`${filterName}_active`]: event.target.checked});
        localStorage.setItem(`filters.${filterName}`, event.target.checked);
        if (typeof window !== "undefined") {
            window.dispatchEvent(new Event("quick-filters"));
        }
    }

    generateRows(filtersNames) {
        return filtersNames.map(name =>
            <tr key={name}>
                <td><input type="checkbox"
                           checked={this.state[`${name}_active`]}
                           onChange={event => this.checkboxChangesHandler(name, event)} /></td>
                <td>{filtersDefinitions[name]}</td>
            </tr>
        );
    }

    render() {
        return (
            <Modal
                {...this.props}
                show="true"
                size="lg"
                aria-labelledby="filters-dialog"
                centered
            >
                <Modal.Header>
                    <Modal.Title id="filters-dialog">
                        ~/filters
                    </Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <Container>
                        <Row>
                            <Col md={6}>
                                <Table borderless size="sm" className="filters-table">
                                    <thead>
                                    <tr>
                                        <th>show</th>
                                        <th>filter</th>
                                    </tr>
                                    </thead>
                                    <tbody>
                                    {this.generateRows(["service_port", "client_address", "min_duration",
                                        "min_bytes", "started_after", "closed_after", "marked"])}
                                    </tbody>
                                </Table>
                            </Col>
                            <Col md={6}>
                                <Table borderless size="sm" className="filters-table">
                                    <thead>
                                    <tr>
                                        <th>show</th>
                                        <th>filter</th>
                                    </tr>
                                    </thead>
                                    <tbody>
                                    {this.generateRows(["matched_rules", "client_port", "max_duration",
                                        "max_bytes", "started_before", "closed_before", "hidden"])}
                                    </tbody>
                                </Table>
                            </Col>

                        </Row>


                    </Container>
                </Modal.Body>
                <Modal.Footer className="dialog-footer">
                    <ButtonField variant="red" bordered onClick={this.props.onHide} name="close" />
                </Modal.Footer>
            </Modal>
        );
    }
}

export default Filters;
