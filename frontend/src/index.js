import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import Filesystem from './filesystem';
import FileView from './fileview';
import FileEdit from './fileedit';
import TableManager from './table_mgr';
import reportWebVitals from './reportWebVitals';
import Admin from './admin.js'
import { Router } from 'react-router';
import { Route } from 'react-router-dom';
import { CookiesProvider, useCookies } from "react-cookie";
import App from './App';
import Login from './login';
import createHistory from 'history/createBrowserHistory';
import TenantAdmin from './tenant_admin';

const history = createHistory();

export default class Frontend extends React.Component {
  constructor(props) {
    super(props);

    this.state = {
      token: props.token,
      tenant: ''
    };
  }

  componentDidMount() {
    if (this.state.token !== '') {
      fetch('http://localhost:8080/auth/token', { headers: new Headers( {'Authorization': 'Bearer ' + this.props.token }) })
        .then(res => {
          return res.text()
        })
        .then(res => {
          if (res.startsWith("Invalid")) {
            this.setState({
              tenant: 'invalid'
            })
            return
          }

          var token = JSON.parse(res)

          this.setState({
            tenant: token.tenant
          })
        })
    }
  }

  render() {
    if (this.state.token !== '' && this.state.tenant === '') {
      return <div>...</div>
    }

    if (typeof this.state.token == 'undefined' || this.state.token === '' || this.state.tenant === 'invalid') {
      return <Login setCookie={this.props.setCookie} history={history}/>
    }

    if (this.state.tenant === '00000000-0000-0000-0000-000000000000') {
      return <Admin token={this.props.token} setCookie={this.props.setCookie} history={history}></Admin>
    }

    return <Router history={history}>
      <Route path="/filesystem/:svc/*" render={(props) => <Filesystem {...props} token={this.props.token} setCookie={this.props.setCookie} history={history}/>}></Route>
      <Route exact path="/filesystem/:svc" render={(props) => <Filesystem {...props} token={this.props.token} setCookie={this.props.setCookie} history={history}/>}></Route>
      <Route path="/fileview/:svc/*" render={(props) => <FileView {...props} token={this.props.token} setCookie={this.props.setCookie} history={history}/>}></Route>
      <Route path="/fileedit/:svc/*" render={(props) => <FileEdit {...props} token={this.props.token} setCookie={this.props.setCookie} history={history}/>}></Route>
      <Route path="/table/:table" render={(props) => <TableManager {...props} token={this.props.token} setCookie={this.props.setCookie} history={history}/>}></Route>
      <Route exact path="/" render={(props) => <TenantAdmin token={this.props.token} setCookie={this.props.setCookie} history={history}/>}></Route>
    </Router>
  }
}

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(
  <CookiesProvider> 
    <App/>
  </CookiesProvider>
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
