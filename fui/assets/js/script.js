class FunctionAPI {
  run(chain, options, data) {
    const chainHeader = chain.join('|');
    const optionsHeader = JSON.stringify(options);
    const headers = new Headers();
    headers.set('X-Btrfaas-Chain', chainHeader);
    headers.set('X-Btrfaas-Options', optionsHeader);
    const fetchOpts = {
      method: 'POST',
      headers: headers,
      body: data,
    };
    return fetch('/api/invoke', fetchOpts).then(res=>res.text());
  }
}

const api = new FunctionAPI();
api.run(['echo-shell'], [{}], 'foobar')
.then(console.log);
