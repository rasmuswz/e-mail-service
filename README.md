
E Mail Service
------------------------------------------------------------

[BitMail Live](https://mail.bitlab.dk) - try logging in. When a new user logs in
for the first time, (s)he is created.

[Click here](#deploy-test-build-get) - to see how the system is built, tested and deployed.

[Background](#background) - a few words about me and web development.


<img src="docs/images/bitmail.png" alt="BitMail" width="250px"/>Concept and Design
--------------
This section describes the conceptual idea and the design choices for a scalable 
and reliable e-mail service.

Concept
---
The concept is a high reliable e-mail sending service. The service is a container that employes several 
Mail Transport Agent Providers (MTA Provider). A single point of e-mail submission is offered by the container. When an 
e-mail is submitted to the service it schedules a provider and forwards the e-mail to it. If a transport fails 
one of the other providers is used as fail-over. If no MTA provider is availble the e-mail service logs that 
it is down and terminates entirely.

To provide a way of sending e-mails the system offers a simple web-mail interface over https. 
The interface allows the user to log in and compose e-mails. Multiple recipient are supported by
separating their address with semi-colon.



<img alt="Dev logo" src="docs/images/devlogo.png" width="80px"/> Design
---
This is a <b>full stack</b> implementation of an e-mail service using several sending providers. See the 
systems component diagram below. 
![System Components Diagram](docs/SystemComponentDiagram.png "E-mail service - System components Diagram")

The solution is implemented in four tiers divided between a server and a client. Three of the tiers run on the server while 
the final teir is a browser on the client. At back of the server we have the MTAServer which is an application implementing the MTA Container. In the middle is the BackEndServer which handles access to the storage. The front of the server, the ClientAPI, accepts HTTPS connections from browser clients. Server applications are written in the [Go](http://www.golang.org) language while the client is written in [Dart](http://www.dartlang.org) compiled to JavaScript.

An MTA-container runs with multiple MTA-provider-components inside to 
offer a unified API for sending (and receiving) emails. Also, it ensures reliability through 
fail-over if any MTA-provider should have a fall-out. The MTA Container uses a Scheduling strategy for chosing 
which MTA provider it shall employ. By default a Round Robin Scheduler is provided. Performance can be optimized
by providing a custom instance of the Scheduler-strategy interface. For example, one could 
implement an adaptive scheduling strategy sending e-mails according to performance stats (e.g. slow MTAs gets scheduled less often).

The system supports three MTAs for sending e-mail: Amazon SES, MailGun, and Mandrill. 
For MailGun the provider is custom and specialized towards using their comprehensive API
including their <b>WebHooks API</b> for getting <b>health information</b> about sent e-mails and 
also their <b>Routes API</b> is used to get notified when e-mails arrive. 
For Amazon SeS there is also a custom provider mainly using their SendEmail function [Rest API](http://docs.aws.amazon.com/ses/latest/DeveloperGuide/sending-email.html). Both Amazon SeS and MailGun offers
Go-libraries to access their services, while Mandrill does not. The Mandrill provider thus creates the Urls and parses 
the repsonse from the Mandrill End-Point. 

The BackEndServer listens for incoming connections from the ClientAPI (the next teir). It uses the storage for user names and other log in related information when authenticating e-mail send requests. The storage is Json-based. Entries are maps from string to string. The storage offers three operations: put, update, and get. Lookup takes a map from string to string and returns all maps in storage that have the given map as a subset. See [jsonstore.go](https://github.com/rasmuswz/e-mail-service/blob/master/goworkspace/src/mail.bitlab.dk/backend/jsonstore.go) The storage container only supports in memory storage for the time being. A final version should include permanent storage like a Oracle <b>MySQL</b> database to implement the jsonstore. A high-performing solution might employ a file based solution in a High-Perf-Distributed-File-Systems like [Hadoop HDFS](https://hadoop.apache.org/docs/stable/hadoop-project-dist/hadoop-hdfs/HdfsUserGuide.html). The storage is used 

The ClientAPI is a Https-webserver which serves the client. Also, Ajax requests are handled by taking appropriate actions with the BackEndServer and the MTAServer. As an example when a users logs in the ClientAPI queries the BackEndServer to authenticate.  

On the client side a Dart-application renders the user interface. It implements the Model-View-Controller pattern having the View defined in [index.html](https://github.com/rasmuswz/e-mail-service/blob/master/dartworkspace/web/index.html), the Controller defined in [main.dart](https://github.com/rasmuswz/e-mail-service/blob/master/dartworkspace/web/main.dart) and a model defined in [mailmodel.dart](https://github.com/rasmuswz/e-mail-service/blob/master/dartworkspace/web/mailmodel.dart). The model takes a strategy for handing communication with the ClientAPI, see  [ClientAPI](https://github.com/rasmuswz/e-mail-service/blob/master/dartworkspace/web/geoconnection.dart) class.

This design has decoupled components in order to scale well. A deployed instance of the system may include several MTAServers, and ClientApis running on different machines. However, to give a consistent experience a common storage is needed, and therefore all ClientAPIs need access to the same BackEndServer. We do provide a Proxy interface for supplying a [ProxyStore](https://github.com/rasmuswz/e-mail-service/blob/master/goworkspace/src/mail.bitlab.dk/backend/jsonstore.go#L270) where the idea is that one instance will have the actual physical storage while other BackEndServer servers could use a Proxy.

Example Deployment
--------------------

To try out the application in practice the domain mail.bitlab.dk has been set up. The domain has been set up to accept email for mail.bitlab.dk and getting mail delivered by MailGun. The domain is supported by two servers: mail0.bitlab.dk hosted here in Aarhus and mail1.bitlab.dk hosted by Amazon AWS in Oregon west. 

The build system uses a Python based tool called  [Fabric](http://www.fabfile.org/). It enables one shell-command to build,
test, commit and deploy new changes to the code base. This gives a short cycle from code to new running features. To get an
idea of how this works you can try it out by logging in at the build-server. The <b>start-up-password</b> is required to log
in, for help with you can also contact <a href="mailto:rwl@cs.au.dk">rwl@cs.au.dk</a>.

<pre>
ssh ubuntu@dev.bitlab.dk<br/>
cd e-mail-service<br/>
ls <br/>
</pre>
Here you will see this repository checked out. The machine is also setup with SSH-Private keys to allow it to deploy
new versions of the software to mail0.bitlab.dk and mail1.bitlab.dk. Try it:

<pre>
fab deploy_bitlab_servers
</pre>

Fabric will execute commands locally to build and test the workspace. Then, it uses ssh to upload files and execute
the necessary ommands on the remote server to stop the running version and replace it. 
To get an overview of what it does see [fabfile.py](https://github.com/rasmuswz/e-mail-service/blob/master/fabfile.py). The deploy function near the bottom lays out what is going on. Old versions are kept on the servers until someone 
logs-in and manually deletes them or reenables them in case of a faulty deployment.

Is it feature complete?
--------------------
No! The central MTA Container component is completed with Failover and Monitoring.
To complete the functionallity of the UI as it is now, we also need to be able to 
   * show incoming e-mails 
   * receive e-mails from more MTAProviders
   * manage different mail-boxes (sent, drafts etc.)

Preparation is already done for some of this in the MTAServer and BackEndServer. For example, 
the MailGun provider accepts incoming mails which are stored by the back-end.

#Deploy, Test, Build, Get
--------------------

See the [Live Demo here](https://mail.bitlab.dk).

To get, build, test and deploy the code yourself you need to follow the steps below.

Getting BitMail
----
Clone this repo.

Preparing your machine
----
You must be running on Linux or OSX and you need the following tools installed :

  * Python 2.17
  * pip (to install Fabric do pip install Fabric)
  * Fabric
  * GoLang-SDK 1.5
  * DartLang-SDK ^1.12.1

Building
----
Change directory to e-mail-server and type <pre>fab build</pre>.

```
e-mail-service$ fab build
```

Fabric builds it all, invoking both the Dart and Go build systems as needed. Notice how <pre>go get</pre> and <pre>pub get</pre> get all the Go and Dart dependencies. 

Testing
----
We only have tests for the components written in Go. To run the Go test suite:
```
e-mail-service$ cd goworkspace && go test
```

Or use Fabric

```
e-mail-service$ fab test
```

Deploying
-----

This will update the production system at mail.bitlab.dk provided you have
got the SSH-RSA private keys to access the servers.
```
fab deploy_bitlab_servers
```
The ssh-keys needed are avaible in the <pre>demo@dev.bitlab.dk:.ssh/ec2key.pem</pre> file on the 
test development environment. As stated above you can log in with <pre>demo@dev.bitlab.dk</pre> using
the <b>start-up-passphrase</b> given in submission note.

Background
-------

I have been working fulltime with software development in C/C++, Java and C#
on the windows platform from 2008 - 2011. My prior experience with Web-development dates back 
to 2005-2007 where he maintained a webpage for the Tutor-association at Aarhus University 
running on a LAMP-box. 

I had no prior experience with the Go language before this exercise, and as such started at 
[A Tour of Go](https://tour.golang.org/welcome/1) Saturday October 31st. 
