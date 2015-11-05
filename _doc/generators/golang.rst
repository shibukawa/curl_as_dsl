Golang
============

Golang generator creates source code that uses ``net/http`` modules. If ``--http2`` is passed, it uses ``golang.org/x/net/http2``.

Only golang generator supports original option ``--awsv2=ACCESS_KEY:SECRET_KEY``. It adds AWS V2 style authentication header. It is good for creating command to access AWS compatible services like `Riak CS <http://basho.com/riak-cloud-storage/>`_.

