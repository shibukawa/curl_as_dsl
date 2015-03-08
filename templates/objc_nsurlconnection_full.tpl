{{ range $key, $_ := .Modules }}#import <{{ $key }}>
{{end}}

BOOL shouldKeepRunning = YES;

@interface HTTPDownloadDelegate : NSObject<NSURLConnectionDelegate> {
    NSMutableData *contents;
}

@end

@implementation HTTPDownloadDelegate

- (void)connection:(NSURLConnection *)connection didReceiveResponse:(NSURLResponse *)response
{
    NSHTTPURLResponse *httpResponse = (NSHTTPURLResponse *)response;
    NSLog(@"Status: %ld", httpResponse.statusCode);
    NSDictionary *headers = httpResponse.allHeaderFields;
    for (id key in headers) {
        NSLog(@"%@: %@", key, [headers objectForKey:key]);
    }
}

- (void)connection:(NSURLConnection *)connection didReceiveData:(NSData *)data
{
    [contents appendData:data];
}

- (void)connectionDidFinishLoading:(NSURLConnection *)connection
{
    NSLog(@"received");
    shouldKeepRunning = NO;
}

@end
{{ .AdditionalDeclaration }}
int main(int argc, char *argv[]) {
    @autoreleasepool {
        {{ .CommonInitialize }}{{ .PrepareBody }}NSMutableURLRequest *request = [NSMutableURLRequest requestWithURL:[NSURL URLWithString:{{ .Url }}]];
{{ .ModifyRequest }}
        HTTPDownloadDelegate *delegate = [[HTTPDownloadDelegate alloc] init];
        
        NSURLConnection *connection = [[NSURLConnection alloc] initWithRequest:request delegate:delegate];

        if (!connection) NSLog(@"failed to create connection");
        
        NSRunLoop *theRL = [NSRunLoop currentRunLoop];
        while (shouldKeepRunning && [theRL runMode:NSDefaultRunLoopMode beforeDate:[NSDate distantFuture]]);
    }
}
