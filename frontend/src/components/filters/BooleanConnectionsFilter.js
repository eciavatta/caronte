import React, {Component} from 'react';
import {withRouter} from "react-router-dom";
import {Redirect} from "react-router";
import CheckField from "../fields/CheckField";

class BooleanConnectionsFilter extends Component {

    constructor(props) {
        super(props);
        this.state = {
            filterActive: "false"
        };

        this.filterChanged = this.filterChanged.bind(this);
        this.needRedirect = false;
    }

    componentDidMount() {
        let params = new URLSearchParams(this.props.location.search);
        this.setState({filterActive: this.toBoolean(params.get(this.props.filterName)).toString()});
    }

    componentDidUpdate(prevProps, prevState, snapshot) {
        let urlParams = new URLSearchParams(this.props.location.search);
        let externalActive = this.toBoolean(urlParams.get(this.props.filterName));
        let filterActive = this.toBoolean(this.state.filterActive);
        // if the filterActive state is changed by another component (and not by filterChanged func) and
        // the query string is not equals at the filterActive state, update the state of the component
        if (this.toBoolean(prevState.filterActive) === filterActive && filterActive !== externalActive) {
            this.setState({filterActive: externalActive.toString()});
        }
    }

    toBoolean(value) {
        return value !== null && value.toLowerCase() === "true";
    }

    filterChanged() {
        this.needRedirect = true;
        this.setState({filterActive: (!this.toBoolean(this.state.filterActive)).toString()});
    }

    render() {
        let redirect = null;
        if (this.needRedirect) {
            let urlParams = new URLSearchParams(this.props.location.search);
            if (this.toBoolean(this.state.filterActive)) {
                urlParams.set(this.props.filterName, "true");
            } else {
                urlParams.delete(this.props.filterName);
            }
            redirect = <Redirect push to={`${this.props.location.pathname}?${urlParams}`} />;

            this.needRedirect = false;
        }

        return (
            <div className="filter" style={{"width": `${this.props.width}px`}}>
                <CheckField checked={this.toBoolean(this.state.filterActive)} name={this.props.filterName}
                            onChange={this.filterChanged} />
                {redirect}
            </div>
        );
    }

}

export default withRouter(BooleanConnectionsFilter);
