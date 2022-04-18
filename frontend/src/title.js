import React from 'react';

import './index.css'
import '../node_modules/bootstrap/dist/css/bootstrap.min.css'

export default class Title extends React.Component {
    handleLogout(t) {
        t.props.setCookie("token", "")
        t.props.history.push("/")
    }

    render() {
        return (
            <header className="d-flex flex-wrap justify-content-center py-3 mb-4 border-bottom">
                <a href="/" className="d-flex align-items-center mb-3 mb-md-0 me-md-auto text-dark text-decoration-none pull-left">
                    <span className="fs-4">{ typeof this.props.detail !== 'undefined' ? this.props.detail : "PaaKS" }</span>
                </a>
                <ul className="nav nav-pills">
                    <li className="nav-item"><a href="#" className="nav-link" onClick={() => this.handleLogout(this)}>Logout</a></li>
                </ul>
            </header>
        );
    }
}