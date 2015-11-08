package test
import (
	"mail.bitlab.dk/model"
	"testing"
	"mail.bitlab.dk/mtacontainer"
	"log"
)


func FreshTestMail(provider mtacontainer.MTAProvider, to string) model.Email {

	var headers map[string][]string = map[string][]string{
		"To": []string{to},
		"From": []string{"BitMail <bitmail@mail.bitlab.dk>"},
		"Subject": []string{"An Email test through " + provider.GetName()},
	}

	var content = "Hi, this is a test, sincerely the BitMail team.";
	return model.NewMail(content, headers);
}


// ---------------------------------------------------------
//
// MTA Failure Strategy Mock
//
// ---------------------------------------------------------
type MockFailureStrategy struct {
	failures       []mtacontainer.EventKind;
	successes      int;
	shouldFailNext bool;
}

func (ths *MockFailureStrategy) setFailNext(failNext bool) {
	ths.shouldFailNext = failNext;
}

func (ths *MockFailureStrategy) Failure(f mtacontainer.EventKind) bool {
	ths.failures = append(ths.failures, f);
	return ths.shouldFailNext;
}

func (ths *MockFailureStrategy) Success() {
	ths.successes += 1;
}

func NewMockFailureStrategy() mtacontainer.FailureStrategy {
	res := new(MockFailureStrategy);
	res.failures = make([]mtacontainer.EventKind, 0);
	res.successes = 0;
	res.shouldFailNext = false;
	return res;
}

// ---------------------------------------------------------
//
// MTA Provider Mock
//
// ---------------------------------------------------------
type MockMTAProvider struct {
	in           chan model.Email;
	out          chan model.Email;
	evt          chan mtacontainer.Event;
	t            *testing.T;
	expectedMail model.Email;
}

func (ths *MockMTAProvider) GetOutgoing() chan model.Email {
	return ths.out;
}

func (ths *MockMTAProvider) GetIncoming() chan model.Email {
	return ths.in;
}

func (ths *MockMTAProvider) Stop() {

}

// Act as a drain eating a message and returning it
// letting the MTA Container believe it has been delivered
func (ths *MockMTAProvider) DoOutgoingEmail() model.Email {
	var mail = <-ths.out;
	log.Println("Got mail");
	if (ths.expectedMail != nil ) {
		log.Println("Not nil");
		if mail != ths.expectedMail {
			print("Error shoudl have happened.")
			ths.t.Error("Expected another email.")
		}
	}
	return mail;
}

// Act a sink add a message as if one was received.
func (ths *MockMTAProvider) DoIncomingEmail(mail model.Email) {
	ths.in <- mail;

}

func (ths *MockMTAProvider) ExpectOutMail(m model.Email) {
	ths.expectedMail = m;
}
func (ths *MockMTAProvider) GetEvent() chan mtacontainer.Event{
	return ths.evt;
}

func (ths *MockMTAProvider) GetName() string {
	return "Mock MTA Provider";
}

func NewMockMTAProvider(t *testing.T) mtacontainer.MTAProvider {
	result := new(MockMTAProvider);
	result.in = make(chan model.Email,5);
	result.out = make(chan model.Email,5);
	result.evt = make(chan mtacontainer.Event);
	result.t = t;
	return result;
}

// ---------------------------------------------------------
//
// Scheduler Mock
//
// ---------------------------------------------------------
type SchedulerMock struct {
	next                 mtacontainer.MTAProvider;
	providers            []mtacontainer.MTAProvider;
	removeProviderResult int;
}

func (ths *SchedulerMock) Schedule() mtacontainer.MTAProvider {
	return ths.next;
}

func (ths *SchedulerMock) SetNextMTA(next *MockMTAProvider) {
	ths.next = next;
}

func (ths *SchedulerMock) GetProviders() []mtacontainer.MTAProvider {
	return ths.providers
}

func (ths *SchedulerMock) SetProviders(providers []mtacontainer.MTAProvider) {
	 ths.providers = providers;
}

func (ths * SchedulerMock) RemoveProviderFromService(provider mtacontainer.MTAProvider) int {
	return ths.removeProviderResult;
}

func (ths * SchedulerMock) setRemoveProviderFromServiceResult(res int) {
	ths.removeProviderResult = res;
}


func NewSchedulerMock() mtacontainer.Scheduler {
	result := new(SchedulerMock);
	result.next = nil;
	return result;
}