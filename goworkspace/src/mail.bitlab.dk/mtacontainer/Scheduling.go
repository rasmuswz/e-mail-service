//
//
//
// Author: Rasmus Winther Zakarias
//
package mtacontainer
// ---------------------------------------------------------
//
// A scheduler shall maintain a set of MTAProviders and upon calling
// {schedule} it returns the next MTAProvider to carry out an
// task according to a implementation specific schedule.
//
// ---------------------------------------------------------
type Scheduler interface {
Schedule() MTAProvider;
GetProviders() []MTAProvider;
RemoveProviderFromService(provider MTAProvider) int;
}



// ------------------------------------------------------
//
// Round robin MTA Provider scheduler
//
// ------------------------------------------------------
type RoundRobinScheduler struct {
	current   int;
	providers []MTAProvider;
}

func (rrs *RoundRobinScheduler) GetProviders() []MTAProvider {
	return rrs.providers;
}

func (rrs *RoundRobinScheduler) Schedule() MTAProvider {
	if (len(rrs.providers) < 1) {
		return nil;
	}
	var result = rrs.providers[rrs.current];
	rrs.current = ((rrs.current+1) % len(rrs.providers));
	return result;
}

//
// If the given provider {mta} is Scheduled by this scheduler
// then remove it from the list of service providers.
//
func (rrs *RoundRobinScheduler) RemoveProviderFromService(mta MTAProvider) int {
	var found bool = false;
	for k := range rrs.providers {
		if mta == rrs.providers[k] {
			found = true;
		}
	}

	if found {
		var i int = 0;
		var newProviders = make([]MTAProvider, len(rrs.providers) - 1);
		for k := range rrs.providers {
			if rrs.providers[k] != mta {
				newProviders[i] = rrs.providers[k];
				i = i + 1;
			}
		}

		if len(newProviders) < 1 {
			rrs.current = 0;
		} else {
			rrs.current = rrs.current % len(newProviders);
		}
		rrs.providers = newProviders;
	}

	return len(rrs.providers);
}

func NewRoundRobinScheduler(providers []MTAProvider) Scheduler {
	if (len(providers) < 1) {
		return nil;
	}

	var rrs = new(RoundRobinScheduler);
	rrs.current = 0;
	rrs.providers = providers;
	return rrs;
}
