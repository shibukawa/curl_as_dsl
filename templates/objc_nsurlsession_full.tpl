{{ range $key, $_ := .Modules }}#import <{{ $key }}>
{{end}}
BOOL shouldKeepRunning = YES;
{{ .AdditionalDeclaration }}
int main(int argc, char *argv[]) {
    @autoreleasepool {
        {{ .CommonInitialize }}{{ .PrepareBody }}NSMutableURLRequest *request = [NSMutableURLRequest requestWithURL:[NSURL URLWithString:{{ .Url }}]];
{{ .ModifyRequest }}
        NSURLSession *session = [NSURLSession sharedSession];
        NSURLSessionDataTask* task = [session dataTaskWithRequest:request completionHandler:^(NSData *data, NSURLResponse *response, NSError *error) {
            NSHTTPURLResponse *httpResponse = (NSHTTPURLResponse *)response;
            NSLog(@"Status: %ld", httpResponse.statusCode);
            NSDictionary *headers = httpResponse.allHeaderFields;
            for (id key in headers) {
                NSLog(@"%@: %@", key, [headers objectForKey:key]);
            }
            if(error == nil) {
                NSString * text = [[NSString alloc] initWithData: data encoding: NSUTF8StringEncoding];
                NSLog(@"Data = %@", text);
            }
            dispatch_sync(dispatch_get_main_queue(), ^(){ shouldKeepRunning = NO; });
        }];
        [task resume];

        NSRunLoop *theRL = [NSRunLoop currentRunLoop];
        while (shouldKeepRunning && [theRL runMode:NSDefaultRunLoopMode beforeDate:[NSDate distantFuture]]);
    }
}
