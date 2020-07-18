import React, {Component} from 'react';
import Typed from 'typed.js';
import './Header.scss';
import {Button} from "react-bootstrap";
import ServicePortFilter from "../components/ServicePortFilter";

class Header extends Component {

    constructor(props) {
        super(props);
        this.state = {
            servicesShow: false
        };
    }

    componentDidMount() {
        const options = {
            strings: ["caronte$ "],
            typeSpeed: 50,
            cursorChar: "❚"
        };
        this.typed = new Typed(this.el, options);
    }

    componentWillUnmount() {
        this.typed.destroy();
    }

    render() {
        return (
            <header className="header container-fluid">
                <div className="row">
                    <div className="col-auto">
                        <h1 className="header-title type-wrap">
                            <span style={{whiteSpace: 'pre'}} ref={(el) => {
                                this.el = el;
                            }}/>
                        </h1>
                    </div>

                    <div className="col-auto">
                        <div className="filters-bar-wrapper">
                            <div className="filters-bar">
                                <ServicePortFilter />
                                {/*<ServicePortFilter name="started_before" default="infinity" />*/}
                                {/*<ServicePortFilter name="started_after" default="-infinity" />*/}

                            </div>
                        </div>
                    </div>

                    <div className="col">
                        <div className="header-buttons">
                            <Button variant="yellow" size="sm">pcaps</Button>
                            <Button variant="blue">rules</Button>
                            <Button variant="red" onClick={this.props.onOpenServices}>
                                services
                            </Button>
                        </div>
                    </div>
                </div>
            </header>
        );
    }
}

export default Header;
