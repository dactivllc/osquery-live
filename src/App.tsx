import React, { Component } from 'react';
import Term from './Term';
import logo from './logo.svg';
import './App.css';

class App extends Component {
  render() {
    return (
      <div className="App">
        <header className="App-header">
          <Term />
        </header>
      </div>
    );
  }
}

export default App;
