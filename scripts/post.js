"use strict";

//wrk.method = "POST";
wrk.body   = "foo=bar&baz=quux";
wrk.header["Content-Type"] = "application/x-www-form-urlencoded";

console.log(wrk)
console.log(wrk.id)
