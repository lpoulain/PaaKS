import React from 'react';
import Title from './title';
import { API } from './api';

import './index.css';
import '../node_modules/bootstrap/dist/css/bootstrap.min.css'

export default class FileEdit extends React.Component {
    constructor(props) {
        super(props);

        this.state = {
            content: '',
            path: typeof this.props.match.params["0"] !== 'undefined'
                ? this.props.match.params.svc + '/' + this.props.match.params["0"]
                : this.props.match.params.svc + '/',
            service: this.props.match.params.svc,
            file: typeof this.props.match.params["0"] !== 'undefined' ? this.props.match.params["0"] : '/'        
        }
    }

    componentDidMount() {
        API.queryText(this, 'filesystem/' + this.state.path, this.props.token, this.dataLoaded)
    }

    dataLoaded(t, results) {
        t.setState({ content: results })
    }

    handleUp(t) {
        var the_arr = t.state.path.split('/');
        the_arr.pop();
        t.props.history.push('/filesystem/' + the_arr.join('/'));
    }

    handleSave(t) {
        API.postForm(t, 'filesystem/' + t.state.path, t.state.content, this.props.token, t.saveCompleted)
    }

    saveCompleted(t, response) {
        t.props.history.push('/fileview/' + t.state.path);
    }

    handleDataChange(t, event) {
        t.setState({"content": event.target.value})
    }

    render() {
        return (
            <div className="container">
                <Title detail={'PaaKS - Service: ' + this.state.service + ', File: /' + this.state.file} setCookie={this.props.setCookie} history={this.props.history}/>
                <div className="col-lg-12">
                    <a href="#" onClick={() => this.handleUp(this)}>Back</a><br/>

                    <textarea className="form-control text-monospace" rows="30" cols="80" value={this.state.content} onChange={(e) => this.handleDataChange(this, e)}/><br/>

                    <button type="button" className="btn btn-primary" onClick={() => this.handleSave(this)}>Save</button>
                </div>
            </div>
        );
    }
}
