import React, {Component} from 'react';
import './App.scss';
import Header from "./Header";
import MainPane from "../components/panels/MainPane";
import Footer from "./Footer";
import {BrowserRouter as Router} from "react-router-dom";
import Filters from "./Filters";
import backend from "../backend";
import ConfigurationPane from "../components/panels/ConfigurationPane";

class App extends Component {

    constructor(props) {
        super(props);

        this.state = {};
    }

    componentDidMount() {
        backend.get("/api/services").then(_ => this.setState({configured: true}));

        setInterval(() => {
            if (document.title.endsWith("❚")) {
                document.title = document.title.slice(0, -1);
            } else {
                document.title += "❚";
            }
        }, 500);
    }

    render() {
        let modal;
        if (this.state.filterWindowOpen && this.state.configured) {
            modal = <Filters onHide={() => this.setState({filterWindowOpen: false})}/>;
        }

        return (
            <div className="main">
                <Router>
                    <div className="main-header">
                        <Header onOpenFilters={() => this.setState({filterWindowOpen: true})} />
                    </div>
                    <div className="main-content">
                        {this.state.configured ? <MainPane /> :
                            <ConfigurationPane onConfigured={() => this.setState({configured: true})} />}
                        {modal}
                    </div>
                    <div className="main-footer">
                        {this.state.configured && <Footer/>}
                    </div>
                </Router>
            </div>
        );
    }
}

export default App;
