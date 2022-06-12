import React from 'react';
import { API } from './api';

import './index.css';
import '../node_modules/bootstrap/dist/css/bootstrap.min.css'

export default class DatabaseManager extends React.Component {
    constructor(props) {
        super(props);

        this.state = {
            tables: []
        }
    }

    componentDidMount() {
        API.queryJson(this, 'db-mgr/tables', this.props.token, this.dataLoaded)
    }

    dataLoaded(t, results) {
        t.setState({ tables: results })
    }

    handleClick(t, table) {
        t.props.history.push('/table/' + table);
    }

    refresh(t, result) {
        alert(result)
        t.componentDidMount();
    }

    render() {
        return (
                <div className="col-lg-12">
                    <div className="panel panel-primary">
                        <div className="panel-heading"></div>
                        <div className="table-responsive">
                            <table className="table table-bordered table-hover table-striped">
                                <thead>
                                    <tr>
                                        <th>Table</th>
                                        <th>Action</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {this.state.tables.map(t => <DatabaseRow key={t} table={t} parent={this}></DatabaseRow>)}
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>
        );
    }
}

class DatabaseRow extends React.Component {
    render() {
        return (
            <tr>
                <td>
                    <a href="#" onClick={() => this.props.parent.handleClick(this.props.parent, this.props.table)}>
                        {this.props.table}
                    </a>
                </td>
                <td>
                    <button type="button" className="btn btn-danger" onClick={() => this.props.parent.handleDelete(this.props.parent, this.props.table)}>Delete</button>
                </td>
            </tr>
        )
    }
}
