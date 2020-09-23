import React, {Component} from 'react';
import './StringField.scss';
import {randomClassName} from "../../utils";

const classNames = require('classnames');

class StringField extends Component {

    constructor(props) {
        super(props);

        this.id = `field-${this.props.name || "noname"}-${randomClassName()}`;
    }

    render() {

        const active = this.props.active || false;
        const invalid = this.props.invalid || false;
        const small = this.props.small || false;
        const inline = this.props.inline || false;
        const name = this.props.name || null;
        const value = this.props.value || "";
        const type = this.props.type || "text";
        const error = this.props.error || null;
        const handler = (e) => {
            if (this.props.onChange) {
                if (e == null) {
                    this.props.onChange("");
                } else {
                    this.props.onChange(e.target.value);
                }
            }
        };

        return (
            <div className={classNames("string-field", {"field-active" : active}, {"field-invalid": invalid},
                {"field-small": small}, {"field-inline": inline})}>
                <div className="field-wrapper">
                    { name &&
                    <div className="field-name">
                        <label id={this.id}>{name}:</label>
                    </div>
                    }
                    <div className="field-input">
                        <div className="field-value">
                            <input type={type} placeholder={this.props.defaultValue} aria-label={name}
                                   aria-describedby={this.id} onChange={handler} value={value} />
                        </div>
                        { value !== "" &&
                        <div className="field-clear">
                            <span onClick={() => handler(null)}>del</span>
                        </div>
                        }
                    </div>
                </div>
                {error &&
                <div className="field-error">
                    error: {error}
                </div>
                }
            </div>
        );
    }
}

export default StringField;
