import axios from 'axios';

export const API = {
    queryJson(t, path, token, callback) {
        const parseFetchResponse = response => response.clone().json().then(text => ({
            json: text,
            meta: response
        }))
        .catch(err => ({json: '', meta: response.clone(), body: response.text()}))
        fetch('http://localhost:8080/' + path, {
            method: 'GET',
            headers: new Headers({
                'Authorization': 'Bearer ' + token,
                'Accept': 'application/json',
                'Content-Type': 'application/json',                      
            })
        })
        .then(parseFetchResponse)
        .then(({ json, meta, body }) => {
            callback(t, json)
        });        
    },

    postJson(t, path, token, body, callback) {
        const parseFetchResponse = response => response.clone().text().then(text => {
            try {
                var json = JSON.parse(text)
                return ({
                    json: json,
                    meta: response
                })
            } catch (error) {
                return ({
                    json: '',
                    meta: response,
                    error: text
                })
            }
        })
//        .catch(err => ({json: '', meta: response.clone(), body: err}))
/*
        const parseFetchResponse = response => response.clone().json().then(text => ({
            json: text,
            meta: response
        }))
        .catch(err => ({json: '', meta: response.clone(), body: err}))
*/
        fetch('http://localhost:8080/' + path, {
            method: 'POST',
            body: body,
            headers: new Headers({
                'Authorization': 'Bearer ' + token,
                'Accept': 'application/json',
                'Content-Type': 'application/json',                      
            })
        })
        .then(parseFetchResponse)
        .then(({ json, meta, error }) => {
            callback(t, json, meta, error)
        });        
    },

    queryText(t, path, token, callback, method = 'GET', body='') {
        fetch('http://localhost:8080/' + path, {
            method: method,
            body: body,
            headers: new Headers({
                'Authorization': 'Bearer ' + token,
                'Accept': 'application/json',
                'Content-Type': 'application/json',                      
            })
        })
            .then(response => response.body)
            .then(body => body.getReader().read())
            .then(body => {
                callback(t, new TextDecoder().decode(body.value), null)
            });
    },

    postForm(t, path, formData, token, callback) {
        axios
        .post("http://localhost:8080/" + path, formData, {
          headers: {
            "Content-Type": "multipart/form-data",
            "Authorization": "Bearer " + token
          }
        })
        .then((response) => {
            callback(t, response, null)
        })
        .catch((e) => {
            callback(t, null, e)
        })
    },

    postMessage(t, path, body, callback) {
        var http = new XMLHttpRequest();
        var url = 'http://localhost:8080/' + path
        http.open('POST', url, true);
    
        //Send the proper header information along with the request
        http.setRequestHeader('Content-type', 'text/plain');
    
        http.onreadystatechange = function() {//Call a function when the state changes.
          if(http.readyState == 4 && http.status == 200) {
            callback(t, http.responseText)
          }
        }
        http.send(body);
    }
}
