What is this?
=====================

This is a code generator. It generates HTTP access codes from `cURL <http://curl.haxx.se/>`_ options.

FAQ
--------

What is the purpose of this tool?
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

I made this tool for:

* creating client code easily.
* creating test code via HTTP easily.
* learning programming languages I didn't know.

Why it doens't support HTTP2 / SPDY?
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

I don't know the programming language that supports HTTP2/SPDY by standard library.

Why it supports only standard libraries? %&#@# library is more popular to access http.
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

I know there are tons of HTTP access libraries. For example `AFNetworking <http://cocoadocs.org/docsets/AFNetworking/2.0.0/index.html>`_ is common in Objective-C programmers.
But using standard library is more stable option. Adding module should be done carefully (do you know npm's complicated dependency hell?).

Bug, I welcome your pull request.

Why cURL?
~~~~~~~~~~~

I was inspired from Google Chrome's dev tool. In its "Network" tab, you can get curl command line text.

Modern langauges provide REPL(Read-Eval-Print-Loop) environment to test code. I think ``cURL`` will be able to be the REPL of RESTful APIs.
I hope all API documents will have ``cURL`` command samples to describe the API behavior like `Riak CS <http://docs.basho.com/riak/latest/dev/references/http/fetch-object/#Simple-Example>`_.

.. image:: _image/devtool.png

Are you RESTful junkie?
~~~~~~~~~~~~~~~~~~~~~~~~~~~

I **HATE** RESTful APIs.

Some RESTful APIs are difficult to use/debug. It needs making complex XML or JSON on the client and add complicated header for authentication.
But it returns just ``400`` when there is some errors. Do you want to drive a car with only one alert lamp? You need to know what is happening on your car,
like gas is empty, tire's air pressure is low or the engine is over heated. Fumbling your way out of ``400`` stresses you and robs your time.

You should use any RPC framework to match impedance between clients and servers like `gRPC <https://github.com/grpc/grpc>`_ if you can.

What is the future plan of this tool?
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

If I need to write some HTTP access code in the language that is not supported now, I will add the code generator for the language.

