import React, {Component} from 'react';
import './Footer.scss';

class Footer extends Component {

    render() {
        return (
            <footer className="footer container-fluid">
                <div className="row">
                    <div className="col-12">
                        <div className="footer-timeline">timeline - <a href="https://github.com/eciavatta/caronte/issues/12">to be implemented</a></div>
                    </div>
                </div>
            </footer>
        );
    }
}

export default Footer;
