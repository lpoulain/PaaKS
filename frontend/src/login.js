import React from 'react';
import Title from './title';
import { API } from './api';

import './index.css';
import '../node_modules/bootstrap/dist/css/bootstrap.min.css'

export default class Login extends React.Component {
    constructor(props) {
        super(props);
    }

    handleLogin(t) {
        API.postMessage(t, "auth/login", document.getElementById("login").value, t.loginDone)
    }

    loginDone(t, results) {
        t.props.setCookie("token", results)
        t.props.history.push("/")
    }

    render() {
        return (
            <div className="modal modal-sheet position-static d-block bg-secondary py-5" tabIndex="-1" role="dialog" id="modalSheet">
                <div className="modal-dialog" role="document">
                    <div className="modal-content rounded-6 shadow">
                    <div className="modal-header border-bottom-0">
                        <h5 className="modal-title">Login</h5>
                    </div>
                    <div className="modal-body py-0">
                        <input type="text" className="w-100 mx-0 mb-2" id="login"></input>
                    </div>
                    <div className="modal-footer flex-column border-top-0">
                        <button type="button" className="btn btn-lg btn-primary w-100 mx-0 mb-2" onClick={() => this.handleLogin(this)}>Login</button>
                    </div>
                    </div>
                </div>
            </div>
        );
    }
}
