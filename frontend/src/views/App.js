import React, {Component} from 'react';
import Header from "./Header";
import './App.scss';
import MainPane from "./MainPane";
import Footer from "./Footer";
import {Route, BrowserRouter as Router, Switch} from "react-router-dom";

class App extends Component {

    render() {
        return (
            <div className="app">
                <Router>
                    <Header/>
                    <Switch>
                        <Route path="/" children={<MainPane/>} />
                        <Route path="/connection/:id" children={<MainPane/>} />
                    </Switch>
                    <Footer/>
                </Router>
            </div>
        );
    }
}

export default App;
