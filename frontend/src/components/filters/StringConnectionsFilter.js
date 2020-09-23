import React, {Component} from 'react';
import {withRouter} from "react-router-dom";
import {Redirect} from "react-router";
import StringField from "../fields/StringField";

class StringConnectionsFilter extends Component {

    constructor(props) {
        super(props);
        this.state = {
            fieldValue: "",
            filterValue: null,
            timeoutHandle: null,
            invalidValue: false
        };
        this.needRedirect = false;
        this.filterChanged = this.filterChanged.bind(this);
    }

    componentDidMount() {
        let params = new URLSearchParams(this.props.location.search);
        this.updateStateFromFilterValue(params.get(this.props.filterName));
    }

    componentDidUpdate(prevProps, prevState, snapshot) {
        let urlParams = new URLSearchParams(this.props.location.search);
        let filterValue = urlParams.get(this.props.filterName);
        if (prevState.filterValue === this.state.filterValue && this.state.filterValue !== filterValue) {
            this.updateStateFromFilterValue(filterValue);
        }
    }

    updateStateFromFilterValue(filterValue) {
        if (filterValue !== null) {
            let fieldValue = filterValue;
            if (typeof this.props.decodeFunc === "function") {
                fieldValue = this.props.decodeFunc(filterValue);
            }
            if (typeof this.props.replaceFunc === "function") {
                fieldValue = this.props.replaceFunc(fieldValue);
            }
            if (this.isValueValid(fieldValue)) {
                this.setState({
                    fieldValue: fieldValue,
                    filterValue: filterValue
                });
            } else {
                this.setState({
                    fieldValue: fieldValue,
                    invalidValue: true
                });
            }
        } else {
            this.setState({fieldValue: "", filterValue: null});
        }
    }

    isValueValid(value) {
        return typeof this.props.validateFunc !== "function" ||
            (typeof this.props.validateFunc === "function" && this.props.validateFunc(value));
    }

    filterChanged(fieldValue) {
        if (this.state.timeoutHandle !== null) {
            clearTimeout(this.state.timeoutHandle);
        }

        if (typeof this.props.replaceFunc === "function") {
            fieldValue = this.props.replaceFunc(fieldValue);
        }

        if (fieldValue === "") {
            this.needRedirect = true;
            this.setState({fieldValue: "", filterValue: null, invalidValue: false});
            return;
        }

        if (this.isValueValid(fieldValue)) {
            let filterValue = fieldValue;
            if (filterValue !== "" && typeof this.props.encodeFunc === "function") {
                filterValue = this.props.encodeFunc(filterValue);
            }

            this.setState({
                fieldValue: fieldValue,
                timeoutHandle: setTimeout(() => {
                    this.needRedirect = true;
                    this.setState({filterValue: filterValue});
                }, 500),
                invalidValue: false
            });
        } else {
            this.needRedirect = true;
            this.setState({
                fieldValue: fieldValue,
                invalidValue: true
            });
        }
    }

    render() {
        let redirect = null;
        if (this.needRedirect) {
            let urlParams = new URLSearchParams(this.props.location.search);
            if (this.state.filterValue !== null) {
                urlParams.set(this.props.filterName, this.state.filterValue);
            } else {
                urlParams.delete(this.props.filterName);
            }
            redirect = <Redirect push to={`${this.props.location.pathname}?${urlParams}`} />;
            this.needRedirect = false;
        }
        let active = this.state.filterValue !== null;

        return (
            <div className="filter" style={{"width": `${this.props.width}px`}}>
                <StringField active={active} invalid={this.state.invalidValue} name={this.props.filterName}
                             defaultValue={this.props.defaultFilterValue} onChange={this.filterChanged}
                             value={this.state.fieldValue} inline={true} small={true} />
                {redirect}
            </div>
        );
    }

}

export default withRouter(StringConnectionsFilter);
