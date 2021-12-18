"use strict";

let baseUrl = wrk.url

let pipeline = {
	token: "",
	authed: false,
	next: function() {
		if (!this.authed) {
			wrk.url = baseUrl + "/auth"
		} else {
			wrk.method = "POST"
			wrk.url = baseUrl + "/foo"
			wrk.body = "foo=bar"
			wrk.header["Authentication"] = "Bearer "+this.token
		}
	}
}

function request() {
	pipeline.next();
}

function response(status, headers, body) {
	if (!pipeline.authed && status == 200) {
		pipeline.authed = true
		pipeline.token = headers["X-Token"]
	}
}