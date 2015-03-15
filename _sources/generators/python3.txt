Python3
==========

It supports Python3 code generation. It uses ``http.client`` module of standard library. If you use Python2, rewrite generated code to use from ``http.client`` to ``urllib2``, or switch to Python3 (recommended).

Restriction
-------------

This code treats all text as string. It will be error when sending binary files by using ``--data-binary`` option or ``-F`` option. You need to fix the code to support binary file.
