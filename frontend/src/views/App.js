import React, {Component} from 'react';
import Header from "./Header";
import './App.scss';
import MainPane from "./MainPane";
import Footer from "./Footer";
import {BrowserRouter as Router, Route, Switch} from "react-router-dom";
import Services from "./Services";

class App extends Component {

    constructor(props) {
        super(props);
        this.state = {
            servicesShow: false
        };
    }

    render() {
        let modal = "";
        if (this.state.servicesShow) {
            modal = <Services onHide={() => this.setState({servicesShow: false})}/>;
        }

        return (
            <div className="app">
                <Router>
                    <Header onOpenServices={() => this.setState({servicesShow: true})}/>
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
