var args = require('system').args;
var server = require('webserver').create();
var webPage = require('webpage');

if(args.length === 1) {
  console.log('You must pass the port number as an argument');
  phantom.exit(1);
}

var PORT = args[1];
console.log('Starting server on port ' + PORT);

var service = server.listen(PORT, function(request, response) {
  console.log('New request')
  //Minimal request validation
  if( request.method != 'POST' ||
      request.headers['Content-Type'] != 'application/json'){
    console.log('Bad request');
    response.statusCode = 400;
    response.write('Bad Request');
    response.close();
    return
  }

  console.log('About to parse post');
  var job = JSON.parse(request.post);

  console.log('Creating page element');
  var page = webPage.create();

  console.log('Loading url: ' + job.URL);
  page.open(job.URL, function(status){
    console.log('Got result from opening');
    if(status !== 'success') {
      console.log('Could not open url')
      response.statusCode = 500;
      response.write('Could not open URL');
      response.close();
      return
    }
    console.log('Success !')
    job['Result'] = page.content;
    response.statusCode = 200;
    response.setHeader('Content-Type', 'application/json')
    response.write(JSON.stringify(job));
    response.close();
  });
});
