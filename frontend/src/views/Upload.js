import React, {Component} from 'react';
import './Upload.scss';
import {Button, ButtonGroup, Col, Container, Form, FormFile, InputGroup, Modal, Row, Table} from "react-bootstrap";
import bsCustomFileInput from 'bs-custom-file-input'
import {createCurlCommand} from '../utils';

class Upload extends Component {

    constructor(props) {
        super(props);

		this.state = {
		  selectedFile: null,
			errors: ""
		};

    }

	onFileChange = event => { 
      this.setState({ selectedFile: event.target.files[0] }); 
     
    }; 

	componentDidMount() {
		bsCustomFileInput.init()
	}

    onFileProcess = () => {
		const formData = new FormData();
		formData.append( 
		  "file", 
		  this.state.selectedFile.name 
		); 
		fetch('/api/pcap/file', {
			method: 'POST',
			body: formData
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
                        /usr/bin/upload
                    </Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <Container>
                        <Row>
							<Form.File
							type="file"
							className="custom-file"
							onChange={this.onFileChange}
							label=".pcap/.pcapng"
							id="custom-file"
							custom
							/>
                        </Row>
						<hr/>
                        <Row>
							<Form.Control
							type="text"
							id="pcap-upload"
							onChange={this.onLocalFileChange}
							placeholder="local .pcap/.pcapng"
							custom
							/>
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

                    <Button variant="blue" onClick={this.onFileProcess}>process_local</Button>
                    <Button variant="green" onClick={this.onFileUpload}>upload</Button>
                    <Button variant="red" onClick={this.props.onHide}>close</Button>
                </Modal.Footer>
            </Modal>
        );
    }
}

export default Upload;
