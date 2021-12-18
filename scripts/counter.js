"use strict";

let counter = 0;
let baseUrl = wrk.url;

function request() {
	wrk.url = baseUrl + "/" + counter;
	wrk.header["X-Counter"] = counter;
	counter++;
}