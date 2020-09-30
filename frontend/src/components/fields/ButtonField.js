import React, {Component} from 'react';
import './ButtonField.scss';
import './common.scss';

const classNames = require('classnames');

class ButtonField extends Component {

    render() {
        const handler = () => {
            if (typeof this.props.onClick === "function") {
                this.props.onClick();
            }
        };

        let buttonClassnames = {
            "button-bordered": this.props.bordered,
        };
        if (this.props.variant) {
            buttonClassnames[`button-variant-${this.props.variant}`] = true;
        }

        let buttonStyle = {};
        if (this.props.color) {
            buttonStyle["backgroundColor"] = this.props.color;
        }
        if (this.props.border) {
            buttonStyle["borderColor"] = this.props.border;
        }
        if (this.props.fullSpan) {
            buttonStyle["width"] = "100%";
        }
        if (this.props.rounded) {
            buttonStyle["borderRadius"] = "3px";
        }
        if (this.props.inline) {
            buttonStyle["marginTop"] = "8px";
        }

        return (
            <div className={classNames( "field", "button-field", {"field-small": this.props.small})}>
                <button type="button" className={classNames(classNames(buttonClassnames))}
                        onClick={handler} style={buttonStyle}>{this.props.name}</button>
            </div>
        );
    }
}

export default ButtonField;
