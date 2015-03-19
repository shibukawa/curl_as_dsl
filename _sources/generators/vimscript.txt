Vim Script
==============

Vim Script generator uses `WebAPI-vim <http://www.vim.org/scripts/script.php?script_id=4019>`_ to speak http protocol. You need to install it.

Restriction
---------------

* WebAPI-vim supports only GET and POST.
* WebAPI-vim can't send any body content when ``get`` is used as a method.
  cURL can do this, but it is not common usage, I think.
* WebAPI-vim supports only one value for each header.
