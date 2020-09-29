import React, {Component} from 'react';
import './Config.scss';
import {Button, ButtonGroup, Col, Container, Form, Modal, Row, Table, ToggleButton} from "react-bootstrap";

class Config extends Component {

    constructor(props) {
        super(props);

        this.state = {
            server_address: "",
            flag_regex: "",
			auth_required: false,
			accounts: {},
			showSignup: false,
			showConfig: true,
			tmpUser:"",
			tmpPass:"",
			tmpConf:"",
			errors:""
        };

        this.serverIpChanged = this.serverIpChanged.bind(this);
        this.flagRegexChanged = this.flagRegexChanged.bind(this);
        this.authRequiredChanged = this.authRequiredChanged.bind(this);
        this.userChanged = this.userChanged.bind(this);
        this.passwdChanged = this.passwdChanged.bind(this);
        this.confirmChanged = this.confirmChanged.bind(this);
    }

    serverIpChanged(event) {
        this.setState({server_address: event.target.value});
    }

    flagRegexChanged(event) {
        this.setState({flag_regex: event.target.value});
    }

    authRequiredChanged() {
        this.setState({auth_required: !this.value});
		this.checked = !this.checked;
		this.value = !this.value;
    }

    userChanged(event) {
        this.setState({tmpUser: event.target.value});
    }

    passwdChanged(event) {
        this.setState({tmpPass: event.target.value});
    }
	
    confirmChanged(event) {
        this.setState({tmpConf: event.target.value});
    }

    setup() {
		const requestOptions = {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ 
			config: { 
				server_address:  this.state.server_address,
				flag_regex: this.state.flag_regex,
				auth_required: this.state.auth_required,
				}, 
				accounts: this.state.accounts  
			})
		};

		fetch('/setup', requestOptions)
			.then(response => {
				if (response.status === 202 ){
					//this.setState({showConfig:false});
					this.props.onHide();
					this.props.onDone();
				} else {
					response.json().then(data => {
						this.setState(
							{errors : data.error.toString()}
						);
					});
				}
			}
		);
		
    }

	signup(){
		if (this.state.tmpPass === this.state.tmpConf){
			const accounts = {...this.state.accounts};
			accounts[this.state.tmpUser] = this.state.tmpPass;
			this.setState({accounts});
			console.log(this.state);
			this.setState({showSignup:false,showConfig:true})
		}
		this.setState({tmpUser : ""});
		this.setState({tmpPass : ""});
		this.setState({tmpConf : ""});
	}

    render() {
		let rows = Object.keys(this.state.accounts).map(u =>
            <tr>
                <td>{u}</td>
            </tr>
        );



        return (
			<>
			<Modal show={this.state.showSignup} size="lg" aria-labelledby="services-dialog" centered >
                <Modal.Header>
                    <Modal.Title id="services-dialog">
                        # passwd
                    </Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <Container>
                        <Row>
                                <Form id="passwd-form">
                                    <Form.Group controlId="username">
                                        <Form.Label>username:</Form.Label>
                                        <Form.Control type="text" onChange={this.userChanged} value={this.state.tmpUser}/>
                                    </Form.Group>

                                    <Form.Group controlId="password">
                                        <Form.Label>password:</Form.Label>
                                        <Form.Control type="password" onChange={this.passwdChanged} value={this.state.tmpPass}/>
                                    </Form.Group>

                                    <Form.Group controlId="confirmPassword">
                                        <Form.Label>confirm password:</Form.Label>
                                        <Form.Control type="password" onChange={this.confirmChanged} value={this.state.tmpConf}/>
                                    </Form.Group>


                                </Form>

                        </Row>

                    </Container>
                </Modal.Body>
                <Modal.Footer className="dialog-footer">
					<Button variant="green" onClick={() => this.signup()}>signup</Button>
                    <Button variant="red" onClick={() => this.setState({showSignup:false,showConfig:true})}>close</Button>
                </Modal.Footer>

			</Modal>
            <Modal
                {...this.props}
                show="true"
                size="lg"
                aria-labelledby="services-dialog"
                centered
            >
                <Modal.Header>
                    <Modal.Title id="services-dialog">
                        ~/.config
                    </Modal.Title>
                </Modal.Header>
                <Modal.Body>
					<div class="blink"><span><b>Warning:</b></span> once the configuration is completed, it cannot be changed unless you reset caronte :(</div> 
					<hr/>
                    <Container>
                        <Row>
                            <Col md={5}>

									<ButtonGroup toggle className="mb-2">
									<ToggleButton
									  type="checkbox"
									  variant="secondary"
									  checked={this.state.auth_required}
									  value={this.state.auth_required}
									  onChange={() => this.authRequiredChanged()}
									>
									  Authentication
									</ToggleButton>
								  </ButtonGroup>

									<Table borderless size="sm" className="users-list">

									<thead>
									<tr>
									<th>users</th>
									</tr>
									</thead>
									<tbody>
									{rows}
									<tr> <td>
									<Button size="sm" onClick={() => this.setState({showSignup:true,showConfig:false})}>new</Button>
									</td> </tr>
									</tbody>
									</Table>



                            </Col>

                            <Col md={7}>

									<Form>
                                    <Form.Group controlId="server_address">
                                        <Form.Label>server_address:</Form.Label>
                                        <Form.Control type="text" onChange={this.serverIpChanged} value={this.state.server_address}/>
                                    </Form.Group>

                                    <Form.Group controlId="flag_regex">
                                        <Form.Label>flag_regex:</Form.Label>
                                        <Form.Control type="text" onChange={this.flagRegexChanged} value={this.state.flag_regex}/>
                                    </Form.Group>

                                </Form>

                            </Col>

                        </Row>
                        <Row>
							<div class="error">
							<b>
								{this.state.errors
									.split('\n').map((item, key) => {
									  return <span key={key}>{item}<br/></span>})
								}
							</b>
							</div>
                        </Row>

                    </Container>
                </Modal.Body>
                <Modal.Footer className="dialog-footer">
					<Button variant="green" onClick={() => this.setup()}>set</Button>
                    <Button variant="red" onClick={this.props.onHide}>close</Button>
                </Modal.Footer>
            </Modal>
			</>
        );
    }
}

export default Config;
