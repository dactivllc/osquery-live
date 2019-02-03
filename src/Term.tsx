import React, { Component } from 'react';
import { Terminal } from 'xterm';
import 'xterm/dist/xterm.css';

export default class Term extends Component {
  xterm: Terminal;
  ref: React.RefObject<HTMLDivElement>;
  socket: WebSocket;

  constructor(props?: any) {
    super(props);
    this.ref = React.createRef<HTMLDivElement>();
    this.xterm = new Terminal({cursorBlink: true});

    const wsHost = process.env.REACT_APP_SERVER_HOST || window.location.host;
    const wsURL = ((window.location.protocol === "https:") ? "wss://" : "ws://") + wsHost + "/shell";
    this.socket = new WebSocket(wsURL);
    this.socket.onmessage = (e) => this.xterm.write(e.data);
    this.xterm.on('data',(data) => this.socket.send(data));
  }

  componentDidMount() {
    if (this.ref.current) {
      this.xterm.open(this.ref.current);
      this.xterm.focus();
    }
  }

  render() {
    return (
      <div ref={this.ref} />
    );
  }
}
