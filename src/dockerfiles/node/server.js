'use strict';
const express = require('express');
const functions = require('./code.js');

// Constants
const PORT = 8080;
const HOST = '0.0.0.0';

// App
const app = express();

Object.keys(functions).map( name => app.all(`/${name}`, functions[name] ))

app.listen(PORT, HOST);
