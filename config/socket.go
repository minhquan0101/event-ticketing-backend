package config

import socketio "github.com/googollee/go-socket.io"

// SocketServer là biến global để controller dùng phát sự kiện realtime
var SocketServer *socketio.Server
