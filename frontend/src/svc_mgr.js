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
        t.setState({ services: results })
    }

    handleClick(t, service) {
        t.props.history.push('/filesystem/' + service);
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
                                    </tr>
                                </thead>
                                <tbody>
                                    {this.state.services.map(svc => <ServiceRow key={svc} name={svc} parent={this}></ServiceRow>)}
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>
            </div>
        );
    }
}

class ServiceRow extends React.Component {
    render() {
        return (
            <tr>
                <td>
                    <a href="#" onClick={() => this.props.parent.handleClick(this.props.parent, this.props.name)}>
                        {this.props.name}
                    </a>
                </td>
            </tr>
        )
    }
}
