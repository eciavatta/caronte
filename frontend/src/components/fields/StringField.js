import React, {Component} from 'react';
import './StringField.scss';

const classNames = require('classnames');

class StringField extends Component {

    render() {
        return (
            <div className={classNames("field", "d-inline-block", {"field-active" : this.props.isActive},
                {"field-invalid": this.props.isInvalid})}>
                <div className="input-group">
                    <div className="field-name-wrapper">
                        <span className="field-name" id={`field-${this.props.name}`}>{this.props.name}:</span>
                    </div>
                    <input placeholder={this.props.defaultValue} aria-label={this.props.name}
                           aria-describedby={`filter-${this.props.name}`} className="field-value"
                           onChange={this.props.onValueChanged} value={this.props.value} />
                </div>

                { this.props.active &&
                <div className="field-clear">
                        <span className="filter-delete-icon" onClick={() => this.props.onValueChanged("")}>del</span>
                </div>
                }
            </div>
        );
    }
}

export default StringField;
