
E Mail Service
------------------------------------------------------------

<div style="width:100%;">
<center><img src="docs/images/geomail.png" alt="GeoMail" width="400px"/></center>
</div>

This is now work-in-progress.

[GeoMail Live](https://mail.bitlab.dk)

[Click here](#deploy-test-build-get)


<img alt="Dev logo" src="docs/images/devlogo.png" width="150px"/> Concept and Design
--------------

The concept intended for this project is GeoMail, the
globally-localized e-mail service. We provide a web-based email
service (e.g. like Gmail) but require that users offer their
location upon log-in. Under the catch-line "<b>You have to share to get
anywhere</b>" we reduce our service for users providing no location
information. In addition to an inbox and the ability to compose
e-mails our service (and this is the new thing) presents mailing lists
of people logged in nearby. An example would be a <b>one mile</b> list
allowing the user to send an email to his contacts (people he has
received or send mails to) logged in with in a range of one mile from the
users current location.

On the technical side we provide a <b>web-mail front-end</b> with a back-end
server <b>abstracting any number</b> of MTA-providers like MailGun, Mandrill,
Amazon SeS and SMTP transports. 
See the [Systems Design](https://github.com/rasmuswz/e-mail-service#system-design) below.
A MTA-container runs with multiple MTA-provider-components inside to 
offers a <b>unified API</b> for <i>sending</i> and <i>receiving</i> emails. Also, it provides ensure reliablity through 
fail-over if ANY MTA-provider should have a fall-out. Wrt. performance the container takes a 
<b>Scheduling</b>-strategy (we provide Round Robin out of the Box) and one could implement an adaptive
scheduling sending e-mails according to performance stats.

We provide a <b>generic</b> SMTP based provider that can interface with PostFix, Gmail, Hotmail or any
other mail-provider using them as MTA-relay for sending e-mails.

We provide a <b>Custom</b> provider for MailGun that is specialized towards using their comprehensive API
including their <b>WebHooks API</b>  for getting <b>health information</b> about sent e-mails and 
also we use their <b>Routes API</b> to get notified when e-mails arrive. 

We provide a <Custom</b> provider for <b>Amazon SeS</b>, TODO(rwz): write me.

To store e-mails for users INBOXes we have a storage container having a <b>REST-based JSon API</b>.
Our storage container only support Oracles <b>MySQL</b> technology for now, however a clear and clean interface
is defined for supporting e.g. a file based solution in a High-Perf-Distributed-File-Systems or other database types.

In this way we intend GeoMail to be scalable and reliable.

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
