import React, {Component} from 'react';
import InputField from "../InputField";

class NumericField extends Component {

    constructor(props) {
        super(props);

        this.state = {
            invalid: false
        };
    }

    render() {
        const handler = (value) => {
            value = value.replace(/[^\d]/gi, '');
            let intValue = 0;
            if (value !== "") {
                intValue = parseInt(value);
            }
            const valid =
                (!this.props.validate || (typeof this.props.validate === "function" && this.props.validate(intValue))) &&
                (!this.props.min || (typeof this.props.min === "number" && intValue >= this.props.min)) &&
                (!this.props.max || (typeof this.props.max === "number" && intValue <= this.props.max));
            this.setState({invalid: !valid});
            if (this.props.onChange) {
                this.props.onChange(intValue);
            }
        };

        return (
            <InputField {...this.props} onChange={handler} initialValue={this.props.initialValue || 0}
                        invalid={this.state.invalid} />
        );
    }

}

export default NumericField;
