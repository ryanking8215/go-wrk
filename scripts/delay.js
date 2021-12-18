"use strict";

let min = 100;
let max = 300;

function delay() { // in ms
	// random between [min, max]
	return Math.floor(Math.random()*(max-min+1)+min);
}
