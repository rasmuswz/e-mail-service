
E Mail Service
------------------------------------------------------------

<div>
<center><img src="docs/images/geomail.png" alt="GeoMail" width="400px"/></center>
</div

This is now work-in-progress.

Want to get going Fast deploying GeoMail, [Click here](#deploy-test-build-get)


![Dev logo](docs/images/devlogo.png) Concept and Design
--------------

The concept intended for this project is GeoMail, the
globally-localized e-mail service. We provide a web-based email
service (e.g. like Gmail) but require that a user offer their
location. Under the catch-line "<b>You have to share to get
anywhere</b>" we reduce our service for users providing no location
information. In addition to an inbox and the ability to compose
e-mails our service (and this is the new thing) presents mailing lists
of people logged in nearby. An example would be a <b>one mile</b> list
allowing the user to send an email to his contacts (people he has
received or send mails to) logged in the range at mile from the
users current location.

On the technical side we provide a web-mail front-end with a back-end
server abstracting any number of MTA-providers like MailGun, Mandrill,
Amazon SeS and SMTP transports. E-mails delivery is ensure by fail-over.
In this way GeoMail is scalable and reliable.

System Design
---------------
![System Components Diagram](docs/SystemComponentDiagram.png "E-mail service - System components Diagram")






#Deploy, Test, Build, Get
--------------------

Well egaer to deploy this project and try it out? [Live Demo](https://mail.bitlab.dk).

Want to do it your self, well we need to Get, Build and optionally run the tests first.

Getting GeoMail
----

Easy just clone this repo.


Building
----
This takes a few easy steps depending on which operating system you are using.

TODO(rwz): Write this section

Testing
----
TODO(rwz): Write this section


Deploying
-----

Easy, this will update the production system at mail.bitlab.dk provided you have
got the SSH-RSA private keys to access the servers.
```
fab deploy
```
TODO(rwz): Write this section.
