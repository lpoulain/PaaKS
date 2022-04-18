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

    queryText(t, path, token, callback) {
        fetch('http://localhost:8080/' + path, {
            method: 'GET',
            headers: new Headers({
                'Authorization': 'Bearer ' + token,
                'Accept': 'application/json',
                'Content-Type': 'application/json',                      
            })
        })
            .then(response => response.body)
            .then(body => body.getReader().read())
            .then(body => {
                callback(t, new TextDecoder().decode(body.value))
            });
    },

    postForm(t, path, body, token, callback) {
        var http = new XMLHttpRequest();
        var url = 'http://localhost:8080/' + path
        var params = 'body=' + encodeURIComponent(body);
        http.open('POST', url, true);
    
        //Send the proper header information along with the request
        http.withCredentials = true
        http.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
        http.setRequestHeader('Authorization', 'Bearer ' + token)
    
        http.onreadystatechange = function() {//Call a function when the state changes.
          if(http.readyState == 4 && http.status == 200) {
            callback(t, http.responseText)
          }
        }
        http.send(params);
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
