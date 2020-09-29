import React, {Component} from 'react';
import Header from "./Header";
import MainPane from "./MainPane";
import Footer from "./Footer";
import {BrowserRouter as Router} from "react-router-dom";
import Services from "./Services";
import Filters from "./Filters";
import Config from "./Config";

class App extends Component {

    constructor(props) {
        super(props);
        this.state = {
            servicesWindowOpen: false,
            filterWindowOpen: false,
            configWindowOpen: false,
            configDone: false
        };

		fetch('/api/services')
		.then(response => {
			if( response.status === 503){
				this.setState({configWindowOpen: true});
			} else if (response.status === 200){
				this.setState({configDone: true});
			}
		});


    }

    render() {
        let modal;
        if (this.state.servicesWindowOpen) {
            modal = <Services onHide={() => this.setState({servicesWindowOpen: false})}/>;
        }
        if (this.state.filterWindowOpen) {
            modal = <Filters onHide={() => this.setState({filterWindowOpen: false})}/>;
        }
        if (this.state.configWindowOpen) {
            modal = <Config onHide={() => this.setState({configWindowOpen: false})}
						onDone={() => this.setState({configDone: true})}/>;
        }

        return (
            <div className="app">
                <Router>
                    <Header onOpenServices={() => this.setState({servicesWindowOpen: true})}
                            onOpenFilters={() => this.setState({filterWindowOpen: true})}
                            onOpenConfig={() => this.setState({configWindowOpen: true})} 
                            onOpenUpload={() => this.setState({uploadWindowOpen: true})} 
							onConfigDone={this.state.configDone}
					/>
                    <MainPane />
                    {modal}
                    <Footer/>
                </Router>

            </div>
        );
    }
}

export default App;
