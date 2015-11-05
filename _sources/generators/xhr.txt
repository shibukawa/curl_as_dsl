XMLHttpRequest of Browsers
===============================

Restriction
-------------

Generated code treats all file contents as binary. So now current generator doesn't support sending file via ``-d @``, ``--data-ascii=@``, ``--data-binary=@``, ``--data-urlencode=@``, ``--F name=<`` and ``--form=name=<``. Use ``-F name=@`` or ``--form=name@`` to send files.

``--http2``, ``--insecure`` are not supported.
