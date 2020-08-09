import React, {Component} from 'react';
import Header from "./Header";
import './App.scss';
import MainPane from "./MainPane";
import Footer from "./Footer";
import {BrowserRouter as Router, Route, Switch} from "react-router-dom";
import Services from "./Services";
import Filters from "./Filters";

class App extends Component {

    constructor(props) {
        super(props);
        this.state = {
            servicesWindowOpen: false,
            filterWindowOpen: false
        };
    }

    render() {
        let modal;
        if (this.state.servicesWindowOpen) {
            modal = <Services onHide={() => this.setState({servicesWindowOpen: false})}/>;
        }
        if (this.state.filterWindowOpen) {
            modal = <Filters onHide={() => this.setState({filterWindowOpen: false})}/>;
        }

        return (
            <div className="app">
                <Router>
                    <Header onOpenServices={() => this.setState({servicesWindowOpen: true})}
                            onOpenFilters={() => this.setState({filterWindowOpen: true})}/>
                    <Switch>
                        <Route path="/connections/:id" children={<MainPane/>}/>
                        <Route path="/" children={<MainPane/>}/>
                    </Switch>
                    {modal}
                    <Footer/>
                </Router>

            </div>
        );
    }
}

export default App;
