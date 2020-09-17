import React, {Component} from 'react';
import './Upload.scss';
import {Button, ButtonGroup, ToggleButton, Col, Container, Form, FormFile, InputGroup, Modal, Row, Table} from "react-bootstrap";
import bsCustomFileInput from 'bs-custom-file-input'
import {createCurlCommand} from '../utils';

class Upload extends Component {

    constructor(props) {
        super(props);

		this.state = {
		  selectedFile: null,
		  removeOriginal: false,
		  flushAll:       false,
			errors: ""
		};

        this.flushAllChanged = this.flushAllChanged.bind(this);
        this.removeOriginalChanged = this.removeOriginalChanged.bind(this);


    }

    flushAllChanged() {
        this.setState({flushAll: !this.value});
		this.checked = !this.checked;
		this.value = !this.value;
    }

    removeOriginalChanged() {
        this.setState({removeOriginal: !this.value});
		this.checked = !this.checked;
		this.value = !this.value;
    }

	onLocalFileChange = event => { 
      this.setState({ selectedFile: event.target.value }); 
     
    }; 

	onFileChange = event => { 
      this.setState({ selectedFile: event.target.files[0] }); 
     
    }; 

	componentDidMount() {
		bsCustomFileInput.init()
	}

    onFileProcess = () => {
		const data = {
		  "file": this.state.selectedFile,
		  "flush_all": this.state.flushAll,
		  "delete_original_file": this.state.removeOriginal};

		fetch('/api/pcap/file', {
			method: 'POST',
			body: JSON.stringify(data)
		})
			.then(response => {
				if (response.status === 202 ){
					this.props.onHide();
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

    onFileUpload = () => {
		const formData = new FormData();
		formData.append( 
		  "file", 
		  this.state.selectedFile, 
		  this.state.selectedFile.name 
		); 
		fetch('/api/pcap/upload', {
			method: 'POST',
			body: formData
		})
			.then(response => {
				if (response.status === 202 ){
					//this.setState({showConfig:false});
					this.props.onHide();
					//this.props.onDone();
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


    render() {

        return (
            <Modal
                {...this.props}
                show="true"
                size="lg"
                aria-labelledby="services-dialog"
                centered
            >
                <Modal.Header>
                    <Modal.Title id="services-dialog">
                        /usr/bin/load_pcap
                    </Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <Container>
                        <Row>
                            <Col>
							--local
                            </Col>
                            <Col>
							--upload
                            </Col>
                        </Row>
                        <Row>
                            <Col>
							<Form.Control
							type="text"
							id="pcap-upload"
							className="custom-file"
							onChange={this.onLocalFileChange}
							placeholder="local .pcap/.pcapng"
							custom
							/>
                            </Col>
                            <Col>
							<Form.File
							type="file"
							className="custom-file"
							onChange={this.onFileChange}
							label=".pcap/.pcapng"
							id="custom-file"
							custom
							/>
                            </Col>
                        </Row>
						<br/>
                        <Row>
                            <Col>
									<ButtonGroup toggle className="mb-2">
									<ToggleButton
									  type="checkbox"
									  variant="secondary"
									  checked={this.state.removeOriginal}
									  value={this.state.removeOriginal}
									  onChange={() => this.removeOriginalChanged()}
									>
									  --remove-original-file
									</ToggleButton>
								  </ButtonGroup>
                            </Col>
                            <Col>
                            </Col>
                        </Row>
                        <Row>
                            <Col>
									<ButtonGroup toggle className="mb-2">
									<ToggleButton
									  type="checkbox"
									  variant="secondary"
									  checked={this.state.flushAll}
									  value={this.state.flushAll}
									  onChange={() => this.flushAllChanged()}
									>
									  --flush-all
									</ToggleButton>
								  </ButtonGroup>
                            </Col>
                            <Col>
                            </Col>
                        </Row>
                        <Row>
                            <Col>
								<br/>
								<Button variant="blue" onClick={this.onFileProcess}>process_local</Button>
                            </Col>
                            <Col>
								<br/>
								<Button variant="green" onClick={this.onFileUpload}>upload</Button>
                            </Col>
                        </Row>
						<hr/>
                        <Row>
							<div class="error">
							<b>
								<br/>
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

                    <Button variant="red" onClick={this.props.onHide}>close</Button>
                </Modal.Footer>
            </Modal>
        );
    }
}

export default Upload;
