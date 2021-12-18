"use strict";

// all config go-wrk supports
wrk.goroutines = 2;
wrk.duration = 100; // in seconds
wrk.timeoutms = 5000;
wrk.redir = true;
wrk.no_comp = true;
wrk.no_keepalive = false;
wrk.skip_verify = false;
wrk.client_cert = "client cert file";
wrk.client_key = "client key file";
wrk.ca_cert = "ca cert file";
wrk.http2 = false;
wrk.method = "GET"; // http method
wrk.host = "";
wrk.header["Content-Type"] = "application/json";
wrk.url = "https://github.com";
wrk.body = "k1=v1&k2=v2";