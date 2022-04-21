import React from 'react';
import Title from './title';
import { API } from './api';

import './index.css';
import '../node_modules/bootstrap/dist/css/bootstrap.min.css'

export default class ServiceManager extends React.Component {
    constructor(props) {
        super(props);

        this.state = {
            services: []
        }
    }

    componentDidMount() {
        API.queryJson(this, 'svc-mgr', this.props.token, this.dataLoaded)
    }

    dataLoaded(t, results) {
        t.setState({ services: Object.entries(results) })
    }

    handleClick(t, service) {
        t.props.history.push('/filesystem/' + service);
    }

    handleCreate(t) {
        API.queryText(this, 'svc-mgr/' + document.getElementById("newService").value, this.props.token, t.refresh, "POST")
    }

    handleDelete(t, service) {
        API.queryText(this, 'svc-mgr/' + service, this.props.token, t.refresh, "DELETE")
    }

    refresh(t, result) {
        alert(result)
        t.componentDidMount();
    }

    render() {
        return (
            <div className="container">
                <Title detail="Welcome to PaaKS (Platform-as-a-Kubernetes-Service)" setCookie={this.props.setCookie} history={this.props.history}/>
                <div className="col-lg-12">
                    <div className="panel panel-primary">
                        <div className="panel-heading"></div>
                        <div className="table-responsive">
                            <table className="table table-bordered table-hover table-striped">
                                <thead>
                                    <tr>
                                        <th>Service</th>
                                        <th>Status</th>
                                        <th>Action</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {this.state.services.map(svc => <ServiceRow key={svc[0]} service={svc} parent={this}></ServiceRow>)}
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>

                <br/>
                <br/>
                <p>Create new service:</p>
                <input className="" id="newService"></input>&nbsp;
                <button type="button" className="btn btn-primary" onClick={() => this.handleCreate(this)}>Create</button>
            </div>
        );
    }
}

class ServiceRow extends React.Component {
    render() {
        return (
            <tr>
                <td>
                    <a href="#" onClick={() => this.props.parent.handleClick(this.props.parent, this.props.service[0])}>
                        {this.props.service[0]}
                    </a>
                </td>
                <td>
                    {Object.entries(this.props.service[1]).map(status => status[0] + ': ' + status[1])}
                </td>
                <td>
                    <button type="button" className="btn btn-danger" onClick={() => this.props.parent.handleDelete(this.props.parent, this.props.service[0])}>Delete</button>
                </td>
            </tr>
        )
    }
}
