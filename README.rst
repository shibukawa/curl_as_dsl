cURL as DSL
===================

This command generates source code from cURL command line options.

Install
---------

.. code-block:: bash

   $ go get -u github.com/shibukawa/curl_as_dsl

Usage
-------

.. code-block:: bash

   $ curl_as_dsl [global options] curl [curl options]

Global Options
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: none

   -t, --target     Code generator name. Now it supports the following generators:

       go, golang         : Golang      (net/http)
       py, python         : Python 3    (http.client)
       node, js.node      : Node.js     (http.request)
       xhr, js.xhr        : Browser     (XMLHttpRequest)
       java               : Java        (java.net.HttpURLConnection)
       objc, objc.session : Objective-C (NSURLSession)
       objc.connection    :             (NSURLConnection)

Supported cURL Options
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: none

   [usage]

          --basic                             Use HTTP Basic Authentication (H)
          --compressed                        Request compressed response (using deflate or gzip)
      -d, --data=DATA                         HTTP POST data (H)
          --data-ascii=DATA                   HTTP POST ASCII data (H)
          --data-binary=DATA                  HTTP POST binary data (H)
          --data-urlencode=DATA               HTTP POST data url encoded (H)
      -G, --get                               Send the -d data with a HTTP GET (H)
      -F, --form=KEY=VALUE                    Specify HTTP multipart POST data (H)
          --form-string=KEY=VALUE             Specify HTTP multipart POST data (H)
      -H, --header=LINE                       Pass custom header LINE to server (H)
      -I, --head                              Show document info only
      -x, --proxy=[PROTOCOL://]HOST[:PORT]    Use proxy on given port
      -e, --referer=                          Referer URL (H)
      -X, --request=COMMAND                   Specify request command to use
          --tr-encoding                       Request compressed transfer encoding (H)
      -T, --upload-file=FILE                  Transfer FILE to destination
          --url=URL                           URL to work with
      -u, --user=USER[:PASSWORD]              Server user and password
      -A, --user-agent=STRING                 User-Agent to send to server (H)

License
---------

MIT License


