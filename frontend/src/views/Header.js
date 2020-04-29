import React, {Component} from 'react';
import Typed from 'typed.js';
import './Header.scss';

class Header extends Component {

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
                    <div className="col">
                        <h1 className="header-title type-wrap">
                            <span style={{ whiteSpace: 'pre' }} ref={(el) => { this.el = el; }} />
                        </h1>
                    </div>
                    <div className="col">
                        <div className="header-buttons">
                            <button className="btn-primary">
                                ➕ pcaps
                            </button>
                            <button className="btn-primary">
                                ➕ rules
                            </button>
                            <button className="btn-primary">
                                ➕ services
                            </button>
                        </div>
                    </div>
                </div>
            </header>
        )
    }
}

export default Header;
