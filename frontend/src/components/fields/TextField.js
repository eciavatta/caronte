import React, {Component} from 'react';
import './TextField.scss';
import {randomClassName} from "../../utils";

const classNames = require('classnames');

class TextField extends Component {

    constructor(props) {
        super(props);

        this.id = `field-${this.props.name || "noname"}-${randomClassName()}`;
    }

    render() {
        const name = this.props.name || null;
        const error = this.props.error || null;
        const rows = this.props.rows || 3;

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
            <div className={classNames("text-field", {"field-active": this.props.active},
                {"field-invalid": this.props.invalid}, {"field-small": this.props.small})}>
                {name && <label htmlFor={this.id}>{name}:</label>}
                <textarea id={this.id} placeholder={this.props.defaultValue} onChange={handler} rows={rows}
                          readOnly={this.props.readonly} value={this.props.value} />
                {error && <div className="field-error">error: {error}</div>}
            </div>
        );
    }
}

export default TextField;
