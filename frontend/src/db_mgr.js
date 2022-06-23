import React from 'react';
import { API } from './api';

import './index.css';
import '../node_modules/bootstrap/dist/css/bootstrap.min.css'

export default class DatabaseManager extends React.Component {
    constructor(props) {
        super(props);

        this.state = {
            tables: [],
            error: ""
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

    handleFileInput(t, e) {
        this.setState({content: e.target.files[0]})
    }

    handleCreate(t) {
        t.setState({error: ''})

        const formData = new FormData();
        formData.append("file", this.state.content);
        
        API.postForm(t, 'db-mgr/tables', formData, this.props.token, t.handleResponse)
    }

    handleDelete(t, table) {
        t.setState({error: ''})
        API.queryText(t, "db-mgr/tables/" + table, this.props.token, t.handleResponse, "DELETE")
    }

    handleResponse(t, response, error) {
        if (error != null) {
            console.log("ERROR " + error.response.status + ' ' + error.response.statusText + ': ' + error.response.data)
            t.setState({error: error.response.data})
        } else {
            API.queryJson(t, 'db-mgr/tables', t.props.token, t.dataLoaded)
        }
    }

    render() {
        return (
            <div>
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

                <div className="col-lg-12">
                <form>
                    <p>Create new table:</p>
                    <p style={{color:'red'}}>{this.state.error}</p>

                    <input className="" id="newTable"></input>&nbsp;
                    <input id="tableDesc" type="file" onChange={(e) => this.handleFileInput(this, e)}></input>
                    <button type="button" className="btn btn-primary" onClick={() => this.handleCreate(this)}>Create</button>
                </form>
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
                    &nbsp;&nbsp;&nbsp;&nbsp;
                    <a href={"http://localhost:8080/db-mgr/tables/" + this.props.table + "/download"}><button type="button" className="btn btn-primary">Download Definition</button></a>
                </td>
            </tr>
        )
    }
}
