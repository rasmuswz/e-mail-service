package test
import (

"testing"
	"mail.bitlab.dk/mtacontainer"
)



func TestMTAProvider_allMock(t *testing.T) {

	var mockProvider  = NewMockMTAProvider(t).(*MockMTAProvider);

	var s *SchedulerMock = NewSchedulerMock().(*SchedulerMock);
	s.SetProviders([]mtacontainer.MTAProvider{mockProvider});
	s.SetNextMTA(mockProvider);
	m := mtacontainer.New(s);

	mail := FreshTestMail(mockProvider,"rwl@cs.au.dk");

	mockProvider.ExpectOutMail(mail);
	m.GetOutgoing() <- mail;

}


func TestMTAProvider_RoundRobinScheduler_test(t *testing.T) {


	var mockP1  = NewMockMTAProvider(t).(*MockMTAProvider);
	var mockP2  = NewMockMTAProvider(t).(*MockMTAProvider);

	var roundRobinScheduler = mtacontainer.NewRoundRobinScheduler([]mtacontainer.MTAProvider{mockP1, mockP2});

	container := mtacontainer.New(roundRobinScheduler);

	mail1 := FreshTestMail(mockP1,"mail1");
	mail2 := FreshTestMail(mockP2,"mail2");
	mail3 := FreshTestMail(mockP2,"mail3");
	mail4 := FreshTestMail(mockP2,"mail4");
	mail5 := FreshTestMail(mockP2,"mail5");
	mail6 := FreshTestMail(mockP2,"mail6");
	mail7 := FreshTestMail(mockP2,"mail7");
	mail8 := FreshTestMail(mockP2,"mail8");
	mail9 := FreshTestMail(mockP2,"mail9");
	mail10 := FreshTestMail(mockP2,"mail10");

	mockP1.ExpectOutMail(mail1);
	mockP2.ExpectOutMail(mail2);
	container.GetOutgoing() <- mail1;
	mockP1.DoOutgoingEmail();
	container.GetOutgoing() <- mail2;
	mockP2.DoOutgoingEmail();

	mockP1.ExpectOutMail(mail3);
	mockP2.ExpectOutMail(mail4);
	container.GetOutgoing() <- mail3;
	mockP1.DoOutgoingEmail();
	container.GetOutgoing() <- mail4;
	mockP2.DoOutgoingEmail();

	mockP1.ExpectOutMail(mail5);
	mockP2.ExpectOutMail(mail6);
	container.GetOutgoing() <- mail5;
	mockP1.DoOutgoingEmail();
	container.GetOutgoing() <- mail6;
	mockP2.DoOutgoingEmail();

	mockP1.ExpectOutMail(mail7);
	mockP2.ExpectOutMail(mail8);
	container.GetOutgoing() <- mail7;
	mockP1.DoOutgoingEmail();
	container.GetOutgoing() <- mail8;
	mockP2.DoOutgoingEmail();


	mockP1.ExpectOutMail(mail9);
	mockP2.ExpectOutMail(mail10);
	container.GetOutgoing() <- mail9;
	mockP1.DoOutgoingEmail();
	container.GetOutgoing() <- mail10;
	mockP2.DoOutgoingEmail();

}