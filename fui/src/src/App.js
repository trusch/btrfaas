import React, { Component } from 'react';
import MuiThemeProvider from 'material-ui/styles/MuiThemeProvider';
import RaisedButton from 'material-ui/RaisedButton';
import FlatButton from 'material-ui/FlatButton';
import {List, ListItem} from 'material-ui/List';
import TextField from 'material-ui/TextField';
import LinearProgress from 'material-ui/LinearProgress';
import {Card, CardHeader, CardText} from 'material-ui/Card';

import logo from './logo.svg';
import './App.css';

class App extends Component {
  constructor(props){
    super(props);
    this.state = {
      functions: [],
      data: '',
      result: '',
      execRunning: false,
    };
  }

  execute() {
    const chainHeader = this.state.functions.map(e=>e.name).join('|');
    const options = [];
    for(let i=0;i<this.state.functions.length;i++){
      options[i] = {}
      if (this.state.functions[i].options) {
        const parts = this.state.functions[i].options.split(' ');
        for(let partIdx=0;partIdx<parts.length; partIdx++) {
          const kv = parts[partIdx].split('=');
          options[i][kv[0]] = kv[1];
        }
      }
    }
    const optionsHeader = JSON.stringify(options);
    const headers = new Headers();
    headers.set('X-Btrfaas-Chain', chainHeader);
    headers.set('X-Btrfaas-Options', optionsHeader);
    const fetchOpts = {
      method: 'POST',
      headers: headers,
      body: this.state.data,
    };
    this.setState({execRunning: true});
    fetch('/api/invoke', fetchOpts).then(res=>res.text())
    .then((result)=>{
      this.setState({result, execRunning: false});
    })
    .catch((e)=>{
      this.setState({error: e, execRunning: false});
    });
  }

  render() {
    return (
      <MuiThemeProvider>
        <div className="App">
          <header className="App-header">
            <img src={logo} className="App-logo" alt="logo" />
            <h1 className="App-title">BtrFaaS Function Chain Editor</h1>
            {this.state.execRunning ? <LinearProgress mode="indeterminate" /> : null}
          </header>
          <TextField value={this.state.data} floatingLabelText={'input data'} onChange={(evt)=>{
            this.setState({
              data: evt.target.value,
            });
          }}/>
          <List>
          {this.state.functions.map((fn,idx)=>{
            return (
              <ListItem key={idx}>
                <TextField value={fn.name} onChange={(evt)=>{
                  const newFunctions = [
                    ...this.state.functions
                  ];
                  newFunctions[idx].name = evt.target.value;
                  this.setState({functions: newFunctions});
                }} floatingLabelText={'function name'}/>
                <TextField value={fn.options} onChange={(evt)=>{
                  const newFunctions = [
                    ...this.state.functions
                  ];
                  newFunctions[idx].options = evt.target.value;
                  this.setState({functions: newFunctions});
                }}floatingLabelText={'options'}/>
                <FlatButton onClick={()=>{
                  const newFunctions = this.state.functions.filter((item,i)=>i!==idx);
                  this.setState({functions: newFunctions})
                }}
                label="delete"/>
              </ListItem>
            );
          })}
          </List>
          <RaisedButton primary={true} style={{margin:'5px'}} onClick={()=>{
            this.setState({functions: [...this.state.functions, {name: 'no-name', options: ''}]})
          }} label="add function"/>
          <RaisedButton primary={true} label="execute" style={{margin:'5px'}} onClick={()=>{this.execute();}}/>
          {this.state.result ? <Card style={{marginLeft: '10%', width:'80%', marginTop:'20px'}}>
            <CardText>
              <h2>Result:</h2>
              {this.state.result}
            </CardText>
          </Card>: null}
        </div>
      </MuiThemeProvider>
    );
  }
}

export default App;
