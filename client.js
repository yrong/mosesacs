var WebSocket = require('ws');

var ws = new WebSocket('ws://localhost:9090/ws');

ws.on('open', function () {
    console.log('connected');
    ws.send(Date.now().toString(), {mask: true});
});

ws.on('close', function () {
    console.log('disconnected');
});

ws.on('message', function (data, flags) {
//    console.log('Roundtrip time: ' + (Date.now() - parseInt(data)) + 'ms', flags);
    console.log(data);
    var s = JSON.parse(data);
    console.log(s);
//    setTimeout(function () {
//        ws.send(Date.now().toString(), {mask: true});
//    }, 500);
});