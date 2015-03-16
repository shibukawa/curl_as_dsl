.. cURL as DSL documentation master file, created by
   sphinx-quickstart on Sun Mar 15 22:02:13 2015.
   You can adapt this file completely to your liking, but it should at least
   contain the root `toctree` directive.

cURL as DSL
=======================================

.. raw:: html

   <style>
   #codeblock {
     white-space: pre;
   }
   #command {
     width: 100%;
     box-sizing: border-box;
     font-family: monospace;
   }
   </style>
   <p>
   <label for="lang">Target:</label>
   <select id="lang">
     <option value="go">Golang</option>
     <option value="python">Python 3</option>
     <option value="php">PHP</option>
     <option value="xhr">JavaScript (XMLHttpRequest)</option>
     <option value="nodejs">JavaScript (node.js)</option>
     <option value="java">Java</option>
     <option value="objc">Objective-C (NSURLSession)</option>
     <option value="objc.connection">Objective-C (NSURLConnection)</option>
   </select>
   </p>
   <p>
   <label for-"command">Curl Command:</label>
   <input type="text" value='curl -F "name=John" -F "photo=@photo.jpg" -H "Accept: text/html" --compressed http://localhost' id="command">
   </p>
   <button id="button">Generate Code</button>
   <pre id="result"><code id="codeblock">Result code is printed here. </code></pre>

Usage
---------

Select target envirionment and type curl command in above text box.
You gets source code that works as same as curl command. Enjoy!

Supported Options
~~~~~~~~~~~~~~~~~~~~~

It doens't support fully options of cURL. It supports only options for http 1.1.

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


Document
------------

.. toctree::
   :maxdepth: 2

   generators/index
   what_is_this
   development

Indices and tables
==================

* :ref:`genindex`
* :ref:`modindex`
* :ref:`search`

