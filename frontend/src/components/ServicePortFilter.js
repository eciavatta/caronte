import React, {Component} from 'react';
import './ServicePortFilter.scss';
import {withRouter} from "react-router-dom";
import {Redirect} from "react-router";

class ServicePortFilter extends Component {

    constructor(props) {
        super(props);
        this.state = {
            servicePort: "",
            servicePortUrl: null,
            timeoutHandle: null
        };

        this.servicePortChanged = this.servicePortChanged.bind(this);
    }

    componentDidMount() {
        let params = new URLSearchParams(this.props.location.search);
        let servicePort = params.get("service_port");
        if (servicePort !== null) {
            this.setState({
                servicePort: servicePort,
                servicePortUrl: servicePort
            });
        }
    }

    servicePortChanged(event) {
        let value = event.target.value.replace(/[^\d]/gi, '');
        if (value.startsWith("0")) {
            return;
        }
        if (value !== "") {
            let port = parseInt(value);
            if (port > 65565) {
                return;
            }
        }

        if (this.state.timeoutHandle !== null) {
            clearTimeout(this.state.timeoutHandle);
        }
        this.setState({
            servicePort: value,
            timeoutHandle: setTimeout(() =>
                this.setState({servicePortUrl: value === "" ? null : value}), 300)
        });
    }

    render() {
        let redirect = null;
        let urlParams = new URLSearchParams(this.props.location.search);
        if (urlParams.get("service_port") !== this.state.servicePortUrl) {
            if (this.state.servicePortUrl !== null) {
                urlParams.set("service_port", this.state.servicePortUrl);
            } else {
                urlParams.delete("service_port");
            }
            redirect = <Redirect push to={`${this.props.location.pathname}?${urlParams}`} />;
        }
        let active = this.state.servicePort !== "";

        return (
            <div className={"filter d-inline-block" + (active ? " filter-active" : "")}
                 style={{"width": "200px"}}>
                <div className="input-group">
                    <div className="filter-name-wrapper">
                        <span className="filter-name" id="filter-service_port">service_port:</span>
                    </div>
                    <input placeholder="all ports" aria-label="service_port" aria-describedby="filter-service_port"
                           className="form-control filter-value" onChange={this.servicePortChanged} value={this.state.servicePort} /></div>

                { active &&
                    <div className="filter-delete">
                        <span className="filter-delete-icon" onClick={() => this.setState({
                            servicePort: "",
                            servicePortUrl: null
                        })}>del</span>
                    </div>
                }

                {redirect}
            </div>
        );
    }

}

export default withRouter(ServicePortFilter);
