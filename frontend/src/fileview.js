import React from 'react';
import Title from './title';
import { API } from './api';

import './index.css';
import '../node_modules/bootstrap/dist/css/bootstrap.min.css'

export default class FileView extends React.Component {
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

    handleEdit(t) {
        t.props.history.push('/fileedit/' + t.state.path);
    }

    render() {
        return (
            <div className="container">
                <Title detail={'PaaKS - Service: ' + this.state.service + ', File: /' + this.state.file} setCookie={this.props.setCookie} history={this.props.history}/>
                <div className="col-lg-12">
                    <a href="#" onClick={() => this.handleUp(this)}>Back</a>

                    <pre style={{borderStyle:'solid', borderWidth:'1px'}}>{this.state.content}</pre>

                    <button type="button" className="btn btn-primary" onClick={() => this.handleEdit(this)}>Edit</button>
                </div>
            </div>
        );
    }
}
