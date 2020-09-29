import React, {Component} from 'react';
import InputField from "../InputField";

class NumericField extends Component {

    constructor(props) {
        super(props);

        this.state = {
            invalid: false
        };
    }

    componentDidUpdate(prevProps, prevState, snapshot) {
        if (prevProps.value !== this.props.value) {
            this.onChange(this.props.value);
        }
    }

    onChange = (value) => {
        value = value.toString().replace(/[^\d]/gi, '');
        let intValue = 0;
        if (value !== "") {
            intValue = parseInt(value);
        }
        const valid =
            (!this.props.validate || (typeof this.props.validate === "function" && this.props.validate(intValue))) &&
            (!this.props.min || (typeof this.props.min === "number" && intValue >= this.props.min)) &&
            (!this.props.max || (typeof this.props.max === "number" && intValue <= this.props.max));
        this.setState({invalid: !valid});
        if (typeof this.props.onChange === "function") {
            this.props.onChange(intValue);
        }
    };

    render() {
        return (
            <InputField {...this.props} onChange={this.onChange} defaultValue={this.props.defaultValue || "0"}
                        invalid={this.state.invalid} />
        );
    }

}

export default NumericField;
