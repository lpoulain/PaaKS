import React from 'react';
import Title from './title';
import { API } from './api';

import './index.css';
import '../node_modules/bootstrap/dist/css/bootstrap.min.css'

export default class ServiceManager extends React.Component {
    constructor(props) {
        super(props);

        this.state = {
            tenants: []
        }
    }

    componentDidMount() {
        API.queryJson(this, 'admin/tenants', this.props.token, this.dataLoaded)
    }

    dataLoaded(t, results) {
        t.setState({ tenants: results })
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
                <Title detail="PaaKS ADMIN" setCookie={this.props.setCookie} history={this.props.history}/>
                <div className="col-lg-12">
                    <div className="panel panel-primary">
                        <div className="panel-heading"></div>
                        <div className="table-responsive">
                            <table className="table table-bordered table-hover table-striped">
                                <thead>
                                    <tr>
                                        <th>Tenant Name</th>
                                        <th>Id</th>
                                        <th>Action</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {this.state.tenants.map(tnt => <TenantRow key={tnt.id} tenant={tnt} parent={this}></TenantRow>)}
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

class TenantRow extends React.Component {
    render() {
        return (
            <tr>
                <td>
                    {this.props.tenant.name}
                </td>
                <td>
                    {this.props.tenant.id}
                </td>
                <td>
                    <button type="button" className="btn btn-danger" onClick={() => this.props.parent.handleDelete(this.props.parent, this.props.tenant.id)}>Delete</button>
                </td>
            </tr>
        )
    }
}
