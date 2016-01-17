var args = require('system').args;
var server = require('webserver').create();
var webPage = require('webpage');

if(args.length === 1) {
  //Must pass port number as argument
  phantom.exit(1);
}

var PORT = args[1];

var service = server.listen(PORT, function(request, response) {
  //Minimal request validation
  if( request.method != 'POST' ||
      request.headers['Content-Type'] != 'application/json'){
    response.statusCode = 400;
    response.write('Bad Request');
    response.close();
    return
  }

  var job = JSON.parse(request.postRaw);

  var page = webPage.create();

  page.open(job.URL, function(status){
    if(status !== 'success') {
      response.statusCode = 500;
      response.write('Could not open URL');
      response.close();
      return
    }
    response.statusCode = 200;
    response.write(page.content);
    response.close();
  });
});
