 // Read request headers for both proxy and target flow
 var headerNames = context.getVariable('request.headers.names');
 var strHeaderNames = String(headerNames);
 var headerList = strHeaderNames.substring(1, strHeaderNames.length - 1).split(new RegExp(', ', 'g'));
 var reqHeaders = {};
 headerList.forEach(function(headerName) {
   reqHeaders[headerName] = context.getVariable('request.header.' + headerName);
 });
 // Read response headers for proxy flow
 headerNames = context.getVariable('response.headers.names');
 strHeaderNames = String(headerNames);
 headerList = strHeaderNames.substring(1, strHeaderNames.length - 1).split(new RegExp(', ', 'g'));
 var resHeaders = {};
 headerList.forEach(function(headerName) {
   resHeaders[headerName] = context.getVariable('response.header.' + headerName);
 });
 
 context.setVariable("apic.reqHeaders", JSON.stringify(JSON.stringify(reqHeaders)));
 context.setVariable("apic.resHeaders", JSON.stringify(JSON.stringify(resHeaders)));