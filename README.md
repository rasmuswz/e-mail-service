
E Mail Service
------------------------------------------------------------

<div style="width:100%;">
<center><img src="docs/images/geomail.png" alt="GeoMail" width="400px"/></center>
</div>

[GeoMail Live](https://mail.bitlab.dk) - Try logging in, if a non existing user logs in
for the first time, (s)he is created.

[Click here](#deploy-test-build-get) - To see how the system is built, tested and deployed.


<img alt="Dev logo" src="docs/images/devlogo.png" width="80px"/> Concept and Design
--------------
This section describes the conceptual idea and the design choices for a scalable 
and reliable service.

Concept
---
The concept intended for this project is GeoMail, the
globally-localized e-mail service. We provide a web-based email
service (e.g. like Gmail) but require that users offer their
location upon log-in. The idea is that "<b>You have to share to get
anywhere</b>" in this case it is your position. In addition to an inbox and the ability to compose
e-mails the service (and this is the new thing) presents mailing lists
of people logged in nearby. An example would be a <b>one mile</b> list
allowing the user to send an email to his contacts (people he has
received or send mails to) logged in with in a range of one mile from the
users current location.

Design
---
This is a full stack implemtation of e-mail service using several sending providers. See the 
systems component diagram below. I provide a <b>web-mail front-end</b> with a back-end
server <b>abstracting any number</b> of MTA-providers like MailGun, Mandrill,
Amazon SeS and SMTP transports. 
![System Components Diagram](docs/SystemComponentDiagram.png "E-mail service - System components Diagram")
A MTA-container runs with multiple MTA-provider-components inside to 
offers a <b>unified API</b> for <i>sending</i> and <i>receiving</i> emails. Also, it ensure reliablity through 
fail-over if ANY MTA-provider should have a fall-out. Wrt. performance it can be optimized by customizing the 
builtin <b>Scheduler</b>-strategy. We provide a default Round Robin out of the Box scheduling strategy. One could 
implement an adaptive scheduling strategy sending e-mails according to performance stats (e.g. slow MTAs gets scheduled less often).

We provide a <b>Custom</b> provider for MailGun that is specialized towards using their comprehensive API
including their <b>WebHooks API</b>  for getting <b>health information</b> about sent e-mails and 
also we use their <b>Routes API</b> to get notified when e-mails arrive. 

We provide a <Custom</b> provider for <b>Amazon SeS</b> using their [Rest API](http://docs.aws.amazon.com/ses/latest/DeveloperGuide/sending-email.html). 

To store e-mails for users INBOXes we define a storage container having a <b>REST-based JSon API</b>. The storage 
contains entries that are of the type: map[string]string or Map<String,String>. An entry is added by providing a map
from string to string. A list of entries can be looked up by providing a matching-map and all records having the keys with
the same values as in the matching map will be return. Finally one can update an entry by given a matching map and a new-values-map and then all entries matching the matching map will have their entries with keys in the new-values-map updated.
See [jsonstore.go](https://github.com/rasmuswz/e-mail-service/blob/master/goworkspace/src/mail.bitlab.dk/backend/jsonstore.go)
Our storage container only supports in memory storage for the time being. A final version should include permanent storage like a Oracle <b>MySQL</b> database to implement the <i>jsonstore</i>. A high-performing solution might employ a file based solution in a High-Perf-Distributed-File-Systems like [Hadoop HDFS](https://hadoop.apache.org/docs/stable/hadoop-project-dist/hadoop-hdfs/HdfsUserGuide.html).

The solution is implemented as three server applications: MTAserver, Backendserver and ClientAPI. The MTA server 
manages the MTA providers and provides a REST API for sending emails. The MTA server also receives e-emails and 
invokes a End-point on the BackEnd-Server when an email has arrived. 

The BackEndServer listens for the MTA Container to deliver email, and stores those received in the jsonstore. Also it listens for the ClientAPI to query INBOX e-mails.

The ClientAPI is a Https-webserver. It serves the build/web folder of a compiled Dart-application which runs on users browsers. The Dart-application uses AJax under the hood to call functionality back on the ClientAPI which in turn forwards requests to the BackEndServer (querying INBOX) and the MTAContainer (sending emails).

The Dart browser application implements the Model-View-Controller pattern having the View defined in [index.html](https://github.com/rasmuswz/e-mail-service/blob/master/dartworkspace/web/index.html), the Controller defined in [main.dart](https://github.com/rasmuswz/e-mail-service/blob/master/dartworkspace/web/main.dart) and a model defined in [mailmodel.dart](https://github.com/rasmuswz/e-mail-service/blob/master/dartworkspace/web/mailmodel.dart). The mailmodel.dadrt takes a strategy for 
handing communication with the ClientAPI, the browser side [ClientAPI](https://github.com/rasmuswz/e-mail-service/blob/master/dartworkspace/web/geoconnection.dart) class.

Example Deployment
--------------------

To try out the application in practice the domain mail.bitlab.dk has been setup. The domain has been setup to accept email for mail.bitlab.dk and getting mail delivered by MailGun, AmazonSes. The domain is supported by two servers: mail0.bitlab.dk hosted here in Aarhus and mail1.bitlab.dk hosted by Amazon AWS in Oregon west. 

To give an idea how the deployment and build system is setup I invite you to take a tour at the build server. The Startup-password is required to login at the server.

<pre>
ssh ubuntu@dev.bitlab.dk<br/>
cd e-mail-service<br/>
ls <br/>
</pre>
Here you will see this respository checked out. This machine is also setup with SSH-Private keys to allow it to deploy
new version of the software to mail0.bitlab.dk and mail1.bitlab.dk. Try it:

<pre>
fab deploy
</pre>

You will see the Python-tool called [Fabric](http://www.fabfile.org/) running the deploy commands once for each server. 
To get an overview of what it does see [fabfile.py](https://github.com/rasmuswz/e-mail-service/blob/master/fabfile.py). The <b>deploy</b> function near the buttom nicely lays out what is going on :-).




Features missing
--------------------
  * The GeoLists needs to be implemented (we do record user locations upon login and store them)
  * Each server instance (we deploy on two servers) 
  *  

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
