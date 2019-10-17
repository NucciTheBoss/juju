// Copyright 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package instancepoller

import (
	"time"

	"github.com/juju/clock"
	"github.com/juju/errors"
	"gopkg.in/juju/names.v3"
	"gopkg.in/juju/worker.v1"
	"gopkg.in/juju/worker.v1/catacomb"

	"github.com/juju/juju/apiserver/params"
	"github.com/juju/juju/core/instance"
	"github.com/juju/juju/core/network"
	"github.com/juju/juju/core/status"
	"github.com/juju/juju/core/watcher"
	"github.com/juju/juju/environs"
	"github.com/juju/juju/environs/context"
	"github.com/juju/juju/environs/instances"
	"github.com/juju/juju/worker/common"
)

//go:generate mockgen -package mocks -destination mocks/mocks_watcher.go github.com/juju/juju/core/watcher StringsWatcher
//go:generate mockgen -package mocks -destination mocks/mocks_instances.go github.com/juju/juju/environs/instances Instance
//go:generate mockgen -package mocks -destination mocks/mocks_cred_api.go github.com/juju/juju/worker/common CredentialAPI
//go:generate mockgen -package mocks -destination mocks/mocks_instancepoller.go github.com/juju/juju/worker/instancepoller Environ,Machine

// ShortPoll and LongPoll hold the polling intervals for the instance
// updater. When a machine has no address or is not started, it will be
// polled at ShortPoll intervals until it does, exponentially backing off
// with an exponent of ShortPollBackoff until a maximum(ish) of LongPoll.
//
// When a machine has an address and is started LongPoll will be used to
// check that the instance address or status has not changed.
var (
	ShortPoll        = 1 * time.Second
	ShortPollBackoff = 2.0
	LongPoll         = 15 * time.Minute
)

// Environ specifies the provider-specific methods needed by the instance
// poller.
type Environ interface {
	Instances(ctx context.ProviderCallContext, ids []instance.Id) ([]instances.Instance, error)
}

// Machine specifies an interface for machine instances processed by the
// instance poller.
type Machine interface {
	Id() string
	InstanceId() (instance.Id, error)
	ProviderAddresses() (network.ProviderAddresses, error)
	SetProviderAddresses(...network.ProviderAddress) error
	InstanceStatus() (params.StatusResult, error)
	SetInstanceStatus(status.Status, string, map[string]interface{}) error
	String() string
	Refresh() error
	Status() (params.StatusResult, error)
	Life() params.Life
	IsManual() (bool, error)
}

// FacadeAPI specifies the api-server methods needed by the instance
// poller.
type FacadeAPI interface {
	WatchModelMachines() (watcher.StringsWatcher, error)
	Machine(tag names.MachineTag) (Machine, error)
}

// Config encapsulates the configuration options for instantiating a new
// instance poller worker.
type Config struct {
	Clock   clock.Clock
	Facade  FacadeAPI
	Environ Environ
	Logger  Logger

	CredentialAPI common.CredentialAPI
}

// Validate checks whether the worker configuration settings are valid.
func (config Config) Validate() error {
	if config.Clock == nil {
		return errors.NotValidf("nil clock.Clock")
	}
	if config.Facade == nil {
		return errors.NotValidf("nil Facade")
	}
	if config.Environ == nil {
		return errors.NotValidf("nil Environ")
	}
	if config.Logger == nil {
		return errors.NotValidf("nil Logger")
	}
	if config.CredentialAPI == nil {
		return errors.NotValidf("nil CredentialAPI")
	}
	return nil
}

type pollGroupType uint8

const (
	shortPollGroup pollGroupType = iota
	longPollGroup
	invalidPollGroup
)

type pollGroupEntry struct {
	m          Machine
	tag        names.MachineTag
	instanceID instance.Id

	shortPollInterval time.Duration
	shortPollAt       time.Time
}

func (e *pollGroupEntry) resetShortPollInterval(clk clock.Clock) {
	e.shortPollInterval = ShortPoll
	e.shortPollAt = clk.Now().Add(e.shortPollInterval)
}

func (e *pollGroupEntry) bumpShortPollInterval(clk clock.Clock) {
	e.shortPollInterval = time.Duration(float64(e.shortPollInterval) * ShortPollBackoff)
	if e.shortPollInterval > LongPoll {
		e.shortPollInterval = LongPoll
	}
	e.shortPollAt = clk.Now().Add(e.shortPollInterval)
}

type updaterWorker struct {
	config   Config
	catacomb catacomb.Catacomb

	pollGroup              [2]map[names.MachineTag]*pollGroupEntry
	instanceIDToGroupEntry map[instance.Id]*pollGroupEntry
	callContext            context.ProviderCallContext

	// Hook function which tests can use to be notified when the worker
	// has processed a full loop iteration.
	loopCompletedHook func()
}

// NewWorker returns a worker that keeps track of
// the machines in the state and polls their instance
// addresses and status periodically to keep them up to date.
func NewWorker(config Config) (worker.Worker, error) {
	if err := config.Validate(); err != nil {
		return nil, errors.Trace(err)
	}
	u := &updaterWorker{
		config: config,
		pollGroup: [2]map[names.MachineTag]*pollGroupEntry{
			make(map[names.MachineTag]*pollGroupEntry),
			make(map[names.MachineTag]*pollGroupEntry),
		},
		instanceIDToGroupEntry: make(map[instance.Id]*pollGroupEntry),
		callContext:            common.NewCloudCallContext(config.CredentialAPI, nil),
	}
	err := catacomb.Invoke(catacomb.Plan{
		Site: &u.catacomb,
		Work: u.loop,
	})
	if err != nil {
		return nil, errors.Trace(err)
	}
	return u, nil
}

// Kill is part of the worker.Worker interface.
func (u *updaterWorker) Kill() {
	u.catacomb.Kill(nil)
}

// Wait is part of the worker.Worker interface.
func (u *updaterWorker) Wait() error {
	return u.catacomb.Wait()
}

func (u *updaterWorker) loop() error {
	watcher, err := u.config.Facade.WatchModelMachines()
	if err != nil {
		return errors.Trace(err)
	}
	if err := u.catacomb.Add(watcher); err != nil {
		return errors.Trace(err)
	}

	shortPollTimer := u.config.Clock.NewTimer(ShortPoll)
	longPollTimer := u.config.Clock.NewTimer(LongPoll)
	defer func() {
		_ = shortPollTimer.Stop()
		_ = longPollTimer.Stop()
	}()

	for {
		select {
		case <-u.catacomb.Dying():
			return u.catacomb.ErrDying()
		case ids, ok := <-watcher.Changes():
			if !ok {
				return errors.New("machines watcher closed")
			}

			for i := range ids {
				tag := names.NewMachineTag(ids[i])
				if err := u.queueMachineForPolling(tag); err != nil {
					return err
				}
			}
		case <-shortPollTimer.Chan():
			if err := u.pollGroupMembers(shortPollGroup); err != nil {
				return err
			}
			shortPollTimer.Reset(ShortPoll)
		case <-longPollTimer.Chan():
			if err := u.pollGroupMembers(longPollGroup); err != nil {
				return err
			}
			longPollTimer.Reset(LongPoll)
		}

		if u.loopCompletedHook != nil {
			u.loopCompletedHook()
		}
	}
}

func (u *updaterWorker) queueMachineForPolling(tag names.MachineTag) error {
	// If we are already polling this machine, check whether it is still alive
	// and remove it from its poll group
	if entry, groupType := u.lookupPolledMachine(tag); entry != nil {
		if err := entry.m.Refresh(); err != nil {
			return errors.Trace(err)
		}
		if entry.m.Life() == params.Dead {
			u.config.Logger.Debugf("removing dead machine %q (instance ID %q)", entry.m, entry.instanceID)
			delete(u.pollGroup[groupType], tag)
			delete(u.instanceIDToGroupEntry, entry.instanceID)
			return nil
		}

		// Something has changed with the machine state. Reset short
		// poll interval for the machine and move it to the short poll
		// group (if not already there) so we immediately poll its
		// status at the next interval.
		entry.resetShortPollInterval(u.config.Clock)
		if groupType == longPollGroup {
			delete(u.pollGroup[longPollGroup], tag)
			u.pollGroup[shortPollGroup][tag] = entry
			u.config.Logger.Debugf("moving machine %q (instance ID %q) to long poll group", entry.m, entry.instanceID)
		}
		return nil
	}

	// Get information about the machine
	m, err := u.config.Facade.Machine(tag)
	if err != nil {
		return errors.Trace(err)
	}

	// We don't poll manual machines, instead we're setting the status to 'running'
	// as we don't have any better information from the provider, see lp:1678981
	isManual, err := m.IsManual()
	if err != nil {
		return errors.Trace(err)
	}

	if isManual {
		machineStatus, err := m.InstanceStatus()
		if err != nil {
			return errors.Trace(err)
		}
		if status.Status(machineStatus.Status) != status.Running {
			if err = m.SetInstanceStatus(status.Running, "Manually provisioned machine", nil); err != nil {
				u.config.Logger.Errorf("cannot set instance status on %q: %v", m, err)
				return err
			}
		}
		return nil
	}

	// Add all new machines to the short poll group and arrange for them to
	// be polled as soon as possible.
	u.appendToShortPollGroup(tag, m)
	return nil
}

func (u *updaterWorker) appendToShortPollGroup(tag names.MachineTag, m Machine) {
	entry := &pollGroupEntry{
		tag: tag,
		m:   m,
	}
	entry.resetShortPollInterval(u.config.Clock)
	u.pollGroup[shortPollGroup][tag] = entry
}

func (u *updaterWorker) lookupPolledMachine(tag names.MachineTag) (*pollGroupEntry, pollGroupType) {
	for groupType, members := range u.pollGroup {
		if found := members[tag]; found != nil {
			return found, pollGroupType(groupType)
		}
	}
	return nil, invalidPollGroup
}

func (u *updaterWorker) pollGroupMembers(groupType pollGroupType) error {
	// Build a list of instance IDs to pass as a query to the provider.
	var instList []instance.Id
	now := u.config.Clock.Now()
	for _, entry := range u.pollGroup[groupType] {
		if groupType == shortPollGroup && now.Before(entry.shortPollAt) {
			continue // we shouldn't poll this entry yet
		}

		if err := u.resolveInstanceID(entry); err != nil {
			if params.IsCodeNotProvisioned(err) {
				// machine not provisioned yet; bump its poll
				// interval and re-try later (or as soon as we
				// get a change for the machine)
				entry.bumpShortPollInterval(u.config.Clock)
				continue
			}
			return errors.Trace(err)
		}

		instList = append(instList, entry.instanceID)
	}

	if len(instList) == 0 {
		return nil
	}

	infoList, err := u.config.Environ.Instances(u.callContext, instList)
	if err != nil && err != environs.ErrPartialInstances {
		return errors.Trace(err)
	}
	for idx, info := range infoList {
		// No details found for this instance. This most probably means
		// that the unit has been killed and we haven't been notified
		// yet. Log the error and keep going.
		if info == nil {
			u.config.Logger.Warningf("unable to retrieve instance information for instance: %q", instList[idx])
			continue
		}

		entry := u.instanceIDToGroupEntry[instList[idx]]
		providerStatus, err := u.processProviderInfo(entry, info)
		if err != nil {
			return errors.Trace(err)
		}

		machineStatus, err := entry.m.Status()
		if err != nil {
			return errors.Trace(err)
		}
		u.maybeSwitchPollGroup(groupType, entry, providerStatus, status.Status(machineStatus.Status))
	}

	return nil
}

func (u *updaterWorker) resolveInstanceID(entry *pollGroupEntry) error {
	if entry.instanceID != "" {
		return nil // already resolved
	}

	instID, err := entry.m.InstanceId()
	if err != nil {
		return errors.Annotate(err, "cannot get machine's instance ID")
	}

	entry.instanceID = instID
	u.instanceIDToGroupEntry[instID] = entry
	return nil
}

func (u *updaterWorker) processProviderInfo(entry *pollGroupEntry, info instances.Instance) (status.Status, error) {
	curStatus, err := entry.m.InstanceStatus()
	if err != nil {
		// This should never occur since the machine is provisioned. If
		// it does occur, report an unknown status to move the machine to
		// the short poll group.
		u.config.Logger.Warningf("cannot get current instance status for machine %v (instance ID %q): %v", entry.m.Id(), entry.instanceID, err)
		return status.Unknown, nil
	}

	// Check for status changes
	providerStatus := info.Status(u.callContext)
	curInstStatus := instance.Status{
		Status:  status.Status(curStatus.Status),
		Message: curStatus.Info,
	}

	if providerStatus != curInstStatus {
		u.config.Logger.Infof("machine %q (instance ID %q) instance status changed from %q to %q", entry.m.Id(), entry.instanceID, curInstStatus, providerStatus)
		if err = entry.m.SetInstanceStatus(providerStatus.Status, providerStatus.Message, nil); err != nil {
			u.config.Logger.Errorf("cannot set instance status on %q: %v", entry.m, err)
			return status.Unknown, errors.Trace(err)
		}

		// If the instance is now running, we should reset the poll
		// interval to make sure we can capture machine status changes
		// as early as possible.
		if providerStatus.Status == status.Running {
			entry.resetShortPollInterval(u.config.Clock)
		}
	}

	// We don't care about dead machines; they will be cleaned up when we
	// process the following machine watcher events.
	if entry.m.Life() == params.Dead {
		return status.Unknown, nil
	}

	// Compare the addreses reported by the provider with the ones recorded
	// at the machine model and trigger an update if required.
	curAddresses, err := entry.m.ProviderAddresses()
	if err != nil {
		return status.Unknown, errors.Trace(err)
	}

	providerAddresses, err := info.Addresses(u.callContext)
	if err != nil {
		return status.Unknown, errors.Trace(err)
	}

	if !addressesEqual(curAddresses, providerAddresses) {
		u.config.Logger.Infof("machine %q (instance ID %q) has new addresses: %v", entry.m.Id(), entry.instanceID, providerAddresses)
		if err := entry.m.SetProviderAddresses(providerAddresses...); err != nil {
			u.config.Logger.Errorf("cannot set addresses on %q: %v", entry.m, err)
			return status.Unknown, errors.Trace(err)
		}
	}

	return providerStatus.Status, nil
}

func (u *updaterWorker) maybeSwitchPollGroup(curGroup pollGroupType, entry *pollGroupEntry, curProviderStatus, curMachineStatus status.Status) {
	if curProviderStatus == status.Allocating || curProviderStatus == status.Pending {
		// Keep the machine in the short poll group until it settles
		entry.bumpShortPollInterval(u.config.Clock)
		return
	}

	machAddrs, _ := entry.m.ProviderAddresses()

	// If the machine is currently in the long poll group and it has an
	// unknown status or suddenly has no network addresses, move it back to
	// the short poll group.
	if curGroup == longPollGroup && (curProviderStatus == status.Unknown || len(machAddrs) == 0) {
		delete(u.pollGroup[longPollGroup], entry.tag)
		u.pollGroup[shortPollGroup][entry.tag] = entry
		entry.resetShortPollInterval(u.config.Clock)
		u.config.Logger.Debugf("moving machine %q (instance ID %q) back to short poll group", entry.m, entry.instanceID)
		return
	}

	// The machine has started and we have at least one address; move to
	// the long poll group
	if len(machAddrs) > 0 && curMachineStatus == status.Started {
		if curGroup == longPollGroup {
			u.config.Logger.Debugf("machine machine %q (instance ID %q) is already in long poll group", entry.m, entry.instanceID)
			return // already in long poll group
		}
		delete(u.pollGroup[shortPollGroup], entry.tag)
		u.pollGroup[longPollGroup][entry.tag] = entry
		u.config.Logger.Debugf("moving machine %q (instance ID %q) to long poll group", entry.m, entry.instanceID)
		return
	}

	// If we are in the short poll group apply exponential backoff to the
	// poll frequency allow time for the machine to boot up.
	if curGroup == shortPollGroup {
		entry.bumpShortPollInterval(u.config.Clock)
	}
}

// addressesEqual compares the addresses of the machine and the instance information.
func addressesEqual(addrs0, addrs1 network.ProviderAddresses) bool {
	if len(addrs0) != len(addrs1) {
		return false
	}

nextAddr:
	for _, a0 := range addrs0 {
		for _, a1 := range addrs1 {
			if a0 == a1 {
				continue nextAddr
			}
		}
		return false
	}

	return true
}
