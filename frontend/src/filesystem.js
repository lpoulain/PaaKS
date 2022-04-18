import React from 'react';
import Title from './title';
import { API } from './api';

import './index.css';
import '../node_modules/bootstrap/dist/css/bootstrap.min.css'

export default class Filesystem extends React.Component {
    constructor(props) {
        super(props);

        this.state = {
            files: [],
            path: typeof this.props.match.params["0"] !== 'undefined'
                ? this.props.match.params.svc + '/' + this.props.match.params["0"]
                : this.props.match.params.svc + '/',
            service: this.props.match.params.svc,
            dir: typeof this.props.match.params["0"] !== 'undefined' ? this.props.match.params["0"] : '/'
        }
    }

    componentDidMount() {
        API.queryJson(this, 'filesystem/' + this.state.path, this.props.token, this.dataLoaded)
    }

    dataLoaded(t, results) {
        t.setState({ files: results })
    }

    handleDir(t, dir) {
        var newPath = t.state.path.endsWith('/') ? t.state.path : t.state.path + '/'
        t.props.history.push('/filesystem/' + newPath + dir);
    }

    handleFile(t, dir) {
        var newPath = t.state.path.endsWith('/') ? t.state.path : t.state.path + '/'
        t.props.history.push('/fileview/' + newPath + dir);
    }

    handleUp(t) {
        var the_arr = t.state.path.split('/');
        the_arr.pop();
        t.props.history.push('/filesystem/' + the_arr.join('/'));
    }

    render() {
        return (
            <div className="container">
                <Title detail={'PaaKS - Service: ' + this.state.service + ', Path: ' + this.state.dir} setCookie={this.props.setCookie} history={this.props.history}/>

                <div className="col-lg-12">
                    { typeof this.props.match.params["0"] !== '' ?
                        <a href="#" onClick={() => this.handleUp(this)}>Up</a>
                        : <span/>
                    }

                    <div className="panel panel-primary">
                        <div className="panel-heading"></div>
                        <div className="table-responsive">
                            <table className="table table-bordered table-hover table-striped">
                                <thead>
                                    <tr>
                                        <th>Name</th>
                                        <th>Directory</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {this.state.files.map(file => file.dir ? <DirRow key={'dir_'+file.name} file={file} parent={this}></DirRow> : <FileRow key={'file_'+file.name} file={file} parent={this}></FileRow>)}
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>
            </div>
        );
    }
}

class FileRow extends React.Component {
    render() {
        return (
            <tr>
                <td>
                    <a href="#" onClick={() => this.props.parent.handleFile(this.props.parent, this.props.file.name)}>
                        {this.props.file.name}
                    </a>
                </td>
                <td>Edit</td>
            </tr>
        )
    }
}

class DirRow extends React.Component {
    render() {
        return (
            <tr>
                <td>
                    <a href="#" onClick={() => this.props.parent.handleDir(this.props.parent, this.props.file.name)}>
                        <b>{this.props.file.name}</b>
                    </a>
                </td>
                <td><b>DIR</b></td>
            </tr>
        )
    }
}