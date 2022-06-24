import React from 'react';
import Title from './title';
import { API } from './api';

import './index.css';
import '../node_modules/bootstrap/dist/css/bootstrap.min.css'

export default class TableManager extends React.Component {
    constructor(props) {
        super(props);

        this.state = {
            columns: [],
            newColumns: [],
            error: ''
        }
    }

    componentDidMount() {
        API.queryJson(this, 'db-mgr/tables/' + this.props.match.params.table, this.props.token, this.dataLoaded)
    }

    dataLoaded(t, results) {
        var i = 1
        results.columns.forEach(r => {
            r.ordinal_position = i++
            r.deleted = false
        })
        t.setState({ columns: results.columns })
    }

    addColumn(t) {
        var columns = t.state.newColumns
        var col = {
            ordinal_position: columns.length + t.state.columns.length + 1,
            name: "",
            data_type: "",
            not_null: false,
            column_default: ""
        }
        columns.push(col)
        t.setState({ newColumns : columns })
    }

    switchColumn(t, col) {
        var columns = t.state.columns
        columns.forEach(c => {
            if (c.ordinal_position == col.ordinal_position) {
                if (typeof c.deleted === 'undefined') {
                    c.deleted = true
                } else {
                    c.deleted = !c.deleted
                }
            }
        });
        t.setState({ columns: columns })
    }

    update(t, field, col, event) {
        var columns = t.state.newColumns
        columns.forEach(c => {
            if (c.ordinal_position == col.ordinal_position) {
                if (field == 'not_null') {
                    c[field] = !c[field]
                } else {
                    c[field] = event.target.value
                }
            }
        });

        t.setState({ newColumns: columns})
    }

    alterTable(t) {
        var cols = []
        t.state.columns.forEach(col => {
            if (col.deleted) {
                cols.push({ 'action': 'DROP', 'name': col.name })
            }
        })
        t.state.newColumns.forEach(col => {
            if (col.not_null && col.column_default !== '') {
                cols.push({ 'action': 'ADD', 'name': col.name, 'data_type': col.data_type, 'not_null': true, 'default': col.column_default })
            } else {
                cols.push({ 'action': 'ADD', 'name': col.name, 'data_type': col.data_type })
            }
        })

        var request = {
            'table': t.props.match.params.table,
            'columns': cols
        }

        console.log(request)
        t.setState({ error: '' })
        API.queryText(t, 'db-mgr/tables/' + t.props.match.params.table, t.props.token, t.alterDone, "PUT", JSON.stringify(request))
    }

    alterDone(t, _, error) {
        if (error !== null) {
            t.setState({ error: error })
        } else {
            t.setState({ newColumns: []})
            API.queryJson(t, 'db-mgr/tables/' + t.props.match.params.table, t.props.token, t.dataLoaded)
        }
    }

    render() {
        return (
            <div className="container">
                <Title detail={'Table "' + this.props.match.params.table + '"'} setCookie={this.props.setCookie} history={this.props.history}/>
                <div className="col-lg-12">
                    <div className="panel panel-primary">
                        <div className="panel-heading"></div>
                        <div className="table-responsive">
                            <table className="table table-bordered table-hover table-striped">
                                <thead>
                                    <tr>
                                        <th>Column Name</th>
                                        <th>Type</th>
                                        <th>Cannot be Null</th>
                                        <th>Default Value</th>
                                        <th>Action</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {this.state.columns.map(col => <Column key={col.name} col={col} parent={this}></Column>)}
                                    {this.state.newColumns.map(col => <NewColumn key={col.ordinal_position} col={col} parent={this}></NewColumn>)}
                                </tbody>
                            </table>
                            <div style={{color:'red'}}>{this.state.error}</div>
                            <div style={{float:'left'}}>
                                <button type="button" className="btn btn-primary" onClick={() => this.addColumn(this)}>Add column</button>
                            </div>
                            <div style={{float:'right'}}>
                                <button type="button" className="btn btn-warning" onClick={() => this.alterTable(this)}>Alter Table</button>
                            </div>
                            <div style={{clear:'both'}}></div>
                        </div>
                    </div>
                </div>
            </div>
        );
    }
}

class Column extends React.Component {
    render() {
        var col = this.props.col
        if (typeof col.deleted !== 'undefined' && col.deleted) {
            return (
                <tr style={{textDecoration:'line-through'}}>
                    <td>{this.props.col.name}</td>
                    <td>{this.props.col.data_type}</td>
                    <td>{this.props.col.not_null ? <span>&#10003;</span> : <span></span>}</td>
                    <td>{this.props.col.column_default}</td>
                    <td>
                        <button type="button" className="btn btn-success" onClick={() => this.props.parent.switchColumn(this.props.parent, this.props.col)}>Undo Remove</button>
                    </td>
                </tr>
            )    
        }
        return (
            <tr>
                <td>{this.props.col.name}</td>
                <td>{this.props.col.data_type}</td>
                <td>{this.props.col.not_null ? <span>&#10003;</span> : <span></span>}</td>
                <td>{this.props.col.column_default}</td>
                <td>
                    <button type="button" className="btn btn-danger" onClick={() => this.props.parent.switchColumn(this.props.parent, this.props.col)}>Remove</button>
                </td>
            </tr>
        )
    }
}

class NewColumn extends React.Component {
    render() {
        return (
            <tr>
                <td><input type="text" value={this.props.col.name} onChange={(event) => this.props.parent.update(this.props.parent, 'name', this.props.col, event)}/></td>
                <td><input type="text" value={this.props.col.data_type} onChange={(event) => this.props.parent.update(this.props.parent, 'data_type', this.props.col, event)}/></td>
                <td><input type="checkbox" value={this.props.col.not_null} onChange={(event) => this.props.parent.update(this.props.parent, 'not_null', this.props.col, event)}/></td>
                <td><input type="text" value={this.props.col.column_default} onChange={(event) => this.props.parent.update(this.props.parent, 'column_default', this.props.col, event)}/></td>
                <td></td>
            </tr>
        )
    }
}
