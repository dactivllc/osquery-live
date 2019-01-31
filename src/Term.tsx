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
    this.socket = new WebSocket(((window.location.protocol === "https:") ? "wss://" : "ws://") + window.location.host + "/shell");
    this.socket.onmessage = (e) => this.xterm.write(e.data);
    this.xterm.on('key',(key, ev) => this.socket.send(key));
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
