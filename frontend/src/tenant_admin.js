import React from 'react';
import Title from './title';
import { API } from './api';
import ServiceManager from './svc_mgr';
import DatabaseManager from './db_mgr';

import './index.css';
import '../node_modules/bootstrap/dist/css/bootstrap.min.css'

export default class TenantAdmin extends React.Component {
    constructor(props) {
        super(props);
    }
    
    render() {
        return (
            <div className="container">
                <Title detail="Welcome to PaaKS (Platform-as-a-Kubernetes-Service)" setCookie={this.props.setCookie} history={this.props.history}/>
                <h2>Services</h2>
                <ServiceManager token={this.props.token} setCookie={this.props.setCookie} history={this.props.history}/>

                <h2>Database</h2>
                <DatabaseManager token={this.props.token} setCookie={this.props.setCookie} history={this.props.history}/>
            </div>
        );
    }
}
