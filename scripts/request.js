"use strcit";

function request() {
	wrk.method = "GET"; // http method
	wrk.host = "";
	wrk.header["Content-Type"] = "application/json";
	wrk.url = "https://github.com";
	wrk.body = "k1=v1&k2=v2";
}