import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import Filesystem from './filesystem';
import FileView from './fileview';
import FileEdit from './fileedit';
import ServiceManager from './svc_mgr';
import reportWebVitals from './reportWebVitals';
import Admin from './admin.js'
import { Router } from 'react-router';
import { Route } from 'react-router-dom';
import { CookiesProvider, useCookies } from "react-cookie";
import App from './App';
import Login from './login';
import createHistory from 'history/createBrowserHistory';

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
        .then(res => res.json())
        .then(token => {
          this.setState({
            tenant: token.tenant
          })
        })
    }
  }

  render() {
    if (typeof this.state.token == 'undefined' || this.state.token === '') {
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
      <Route exact path="/" render={(props) => <ServiceManager token={this.props.token} setCookie={this.props.setCookie} history={history}/>}></Route>
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
