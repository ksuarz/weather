weather
=======

Assignment 8 for Distributed Systems.

Starting the Server
-------------------
To start the server, first compile it:

    $ go build weather.go

Then, run the output executable:

    $ ./weather

Making Requests
---------------
The default port for this application is `8080`; you can interact with it using
a REST-like interface:

    $ wget localhost:8080/jersey_city
