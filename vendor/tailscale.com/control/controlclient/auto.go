// Copyright (c) Tailscale Inc & AUTHORS
// SPDX-License-Identifier: BSD-3-Clause

package controlclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"tailscale.com/health"
	"tailscale.com/logtail/backoff"
	"tailscale.com/net/sockstats"
	"tailscale.com/tailcfg"
	"tailscale.com/tstime"
	"tailscale.com/types/empty"
	"tailscale.com/types/key"
	"tailscale.com/types/logger"
	"tailscale.com/types/netmap"
	"tailscale.com/types/persist"
	"tailscale.com/types/ptr"
	"tailscale.com/types/structs"
)

type LoginGoal struct {
	_               structs.Incomparable
	wantLoggedIn    bool                 // true if we *want* to be logged in
	token           *tailcfg.Oauth2Token // oauth token to use when logging in
	flags           LoginFlags           // flags to use when logging in
	url             string               // auth url that needs to be visited
	loggedOutResult chan<- error
}

func (g *LoginGoal) sendLogoutError(err error) {
	if g.loggedOutResult == nil {
		return
	}
	select {
	case g.loggedOutResult <- err:
	default:
	}
}

var _ Client = (*Auto)(nil)

// waitUnpause waits until the client is unpaused then returns. It only
// returns an error if the client is closed.
func (c *Auto) waitUnpause(routineLogName string) error {
	c.mu.Lock()
	if !c.paused {
		c.mu.Unlock()
		return nil
	}
	unpaused := c.unpausedChanLocked()
	c.mu.Unlock()
	c.logf("%s: awaiting unpause", routineLogName)
	select {
	case <-unpaused:
		c.logf("%s: unpaused", routineLogName)
		return nil
	case <-c.quit:
		return errors.New("quit")
	}
}

// updateRoutine is responsible for informing the server of worthy changes to
// our local state. It runs in its own goroutine.
func (c *Auto) updateRoutine() {
	defer close(c.updateDone)
	bo := backoff.NewBackoff("updateRoutine", c.logf, 30*time.Second)

	// lastUpdateGenInformed is the value of lastUpdateAt that we've successfully
	// informed the server of.
	var lastUpdateGenInformed updateGen

	for {
		if err := c.waitUnpause("updateRoutine"); err != nil {
			c.logf("updateRoutine: exiting")
			return
		}
		c.mu.Lock()
		gen := c.lastUpdateGen
		ctx := c.mapCtx
		needUpdate := gen > 0 && gen != lastUpdateGenInformed && c.loggedIn
		c.mu.Unlock()

		if needUpdate {
			select {
			case <-c.quit:
				c.logf("updateRoutine: exiting")
				return
			default:
			}
		} else {
			// Nothing to do, wait for a signal.
			select {
			case <-c.quit:
				c.logf("updateRoutine: exiting")
				return
			case <-c.updateCh:
				continue
			}
		}

		t0 := c.clock.Now()
		err := c.direct.SendUpdate(ctx)
		d := time.Since(t0).Round(time.Millisecond)
		if err != nil {
			if ctx.Err() == nil {
				c.direct.logf("lite map update error after %v: %v", d, err)
			}
			bo.BackOff(ctx, err)
			continue
		}
		bo.BackOff(ctx, nil)
		c.direct.logf("[v1] successful lite map update in %v", d)

		lastUpdateGenInformed = gen
	}
}

// atomicGen is an atomic int64 generator. It is used to generate monotonically
// increasing numbers for updateGen.
var atomicGen atomic.Int64

func nextUpdateGen() updateGen {
	return updateGen(atomicGen.Add(1))
}

// updateGen is a monotonically increasing number that represents a particular
// update to the local state.
type updateGen int64

// Auto connects to a tailcontrol server for a node.
// It's a concrete implementation of the Client interface.
type Auto struct {
	direct     *Direct // our interface to the server APIs
	clock      tstime.Clock
	logf       logger.Logf
	expiry     *time.Time
	closed     bool
	updateCh   chan struct{} // readable when we should inform the server of a change
	newMapCh   chan struct{} // readable when we must restart a map request
	statusFunc func(Status)  // called to update Client status; always non-nil

	unregisterHealthWatch func()

	mu sync.Mutex // mutex guards the following fields

	// lastUpdateGen is the gen of last update we had an update worth sending to
	// the server.
	lastUpdateGen updateGen

	paused         bool // whether we should stop making HTTP requests
	unpauseWaiters []chan struct{}
	loggedIn       bool       // true if currently logged in
	loginGoal      *LoginGoal // non-nil if some login activity is desired
	synced         bool       // true if our netmap is up-to-date
	inSendStatus   int        // number of sendStatus calls currently in progress
	state          State

	authCtx    context.Context // context used for auth requests
	mapCtx     context.Context // context used for netmap and update requests
	authCancel func()          // cancel authCtx
	mapCancel  func()          // cancel mapCtx
	quit       chan struct{}   // when closed, goroutines should all exit
	authDone   chan struct{}   // when closed, authRoutine is done
	mapDone    chan struct{}   // when closed, mapRoutine is done
	updateDone chan struct{}   // when closed, updateRoutine is done
}

// New creates and starts a new Auto.
func New(opts Options) (*Auto, error) {
	c, err := NewNoStart(opts)
	if c != nil {
		c.Start()
	}
	return c, err
}

// NewNoStart creates a new Auto, but without calling Start on it.
func NewNoStart(opts Options) (_ *Auto, err error) {
	direct, err := NewDirect(opts)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			direct.Close()
		}
	}()

	if opts.Status == nil {
		return nil, errors.New("missing required Options.Status")
	}
	if opts.Logf == nil {
		opts.Logf = func(fmt string, args ...any) {}
	}
	if opts.Clock == nil {
		opts.Clock = tstime.StdClock{}
	}
	c := &Auto{
		direct:     direct,
		clock:      opts.Clock,
		logf:       opts.Logf,
		updateCh:   make(chan struct{}, 1),
		newMapCh:   make(chan struct{}, 1),
		quit:       make(chan struct{}),
		authDone:   make(chan struct{}),
		mapDone:    make(chan struct{}),
		updateDone: make(chan struct{}),
		statusFunc: opts.Status,
	}
	c.authCtx, c.authCancel = context.WithCancel(context.Background())
	c.authCtx = sockstats.WithSockStats(c.authCtx, sockstats.LabelControlClientAuto, opts.Logf)

	c.mapCtx, c.mapCancel = context.WithCancel(context.Background())
	c.mapCtx = sockstats.WithSockStats(c.mapCtx, sockstats.LabelControlClientAuto, opts.Logf)

	c.unregisterHealthWatch = health.RegisterWatcher(direct.ReportHealthChange)
	return c, nil

}

// SetPaused controls whether HTTP activity should be paused.
//
// The client can be paused and unpaused repeatedly, unlike Start and Shutdown, which can only be used once.
func (c *Auto) SetPaused(paused bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if paused == c.paused {
		return
	}
	c.logf("setPaused(%v)", paused)
	c.paused = paused
	if paused {
		// Only cancel the map routine. (The auth routine isn't expensive
		// so it's fine to keep it running.)
		c.cancelMapLocked()
	} else {
		for _, ch := range c.unpauseWaiters {
			close(ch)
		}
		c.unpauseWaiters = nil
	}
}

// Start starts the client's goroutines.
//
// It should only be called for clients created by NewNoStart.
func (c *Auto) Start() {
	go c.authRoutine()
	go c.mapRoutine()
	go c.updateRoutine()
}

// updateControl sends a new OmitPeers, non-streaming map request (to just send
// Hostinfo/Netinfo/Endpoints info, while keeping an existing streaming response
// open).
//
// It should be called whenever there's something new to tell the server.
func (c *Auto) updateControl() {
	gen := nextUpdateGen()
	c.mu.Lock()
	if gen < c.lastUpdateGen {
		// This update is out of date.
		c.mu.Unlock()
		return
	}
	c.lastUpdateGen = gen
	c.mu.Unlock()

	select {
	case c.updateCh <- struct{}{}:
	default:
	}
}

func (c *Auto) cancelAuth() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.authCancel != nil {
		c.authCancel()
	}
	if !c.closed {
		c.authCtx, c.authCancel = context.WithCancel(context.Background())
		c.authCtx = sockstats.WithSockStats(c.authCtx, sockstats.LabelControlClientAuto, c.logf)
	}
}

// cancelMapLocked is like cancelMap, but assumes the caller holds c.mu.
func (c *Auto) cancelMapLocked() {
	if c.mapCancel != nil {
		c.mapCancel()
	}
	if !c.closed {
		c.mapCtx, c.mapCancel = context.WithCancel(context.Background())
		c.mapCtx = sockstats.WithSockStats(c.mapCtx, sockstats.LabelControlClientAuto, c.logf)
	}
}

// cancelMap cancels the existing mapPoll and liteUpdates.
func (c *Auto) cancelMap() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cancelMapLocked()
}

// restartMap cancels the existing mapPoll and liteUpdates, and then starts a
// new one.
func (c *Auto) restartMap() {
	c.mu.Lock()
	c.cancelMapLocked()
	synced := c.synced
	c.mu.Unlock()

	c.logf("[v1] restartMap: synced=%v", synced)

	select {
	case c.newMapCh <- struct{}{}:
		c.logf("[v1] restartMap: wrote to channel")
	default:
		// if channel write failed, then there was already
		// an outstanding newMapCh request. One is enough,
		// since it'll always use the latest endpoints.
		c.logf("[v1] restartMap: channel was full")
	}
	c.updateControl()
}

func (c *Auto) authRoutine() {
	defer close(c.authDone)
	bo := backoff.NewBackoff("authRoutine", c.logf, 30*time.Second)

	for {
		c.mu.Lock()
		goal := c.loginGoal
		ctx := c.authCtx
		if goal != nil {
			c.logf("[v1] authRoutine: %s; wantLoggedIn=%v", c.state, goal.wantLoggedIn)
		} else {
			c.logf("[v1] authRoutine: %s; goal=nil paused=%v", c.state, c.paused)
		}
		c.mu.Unlock()

		select {
		case <-c.quit:
			c.logf("[v1] authRoutine: quit")
			return
		default:
		}

		report := func(err error, msg string) {
			c.logf("[v1] %s: %v", msg, err)
			// don't send status updates for context errors,
			// since context cancelation is always on purpose.
			if ctx.Err() == nil {
				c.sendStatus("authRoutine-report", err, "", nil)
			}
		}

		if goal == nil {
			health.SetAuthRoutineInError(nil)
			// Wait for user to Login or Logout.
			<-ctx.Done()
			c.logf("[v1] authRoutine: context done.")
			continue
		}

		if !goal.wantLoggedIn {
			health.SetAuthRoutineInError(nil)
			err := c.direct.TryLogout(ctx)
			goal.sendLogoutError(err)
			if err != nil {
				report(err, "TryLogout")
				bo.BackOff(ctx, err)
				continue
			}

			// success
			c.mu.Lock()
			c.loggedIn = false
			c.loginGoal = nil
			c.state = StateNotAuthenticated
			c.synced = false
			c.mu.Unlock()

			c.sendStatus("authRoutine-wantout", nil, "", nil)
			bo.BackOff(ctx, nil)
		} else { // ie. goal.wantLoggedIn
			c.mu.Lock()
			if goal.url != "" {
				c.state = StateURLVisitRequired
			} else {
				c.state = StateAuthenticating
			}
			c.mu.Unlock()

			var url string
			var err error
			var f string
			if goal.url != "" {
				url, err = c.direct.WaitLoginURL(ctx, goal.url)
				f = "WaitLoginURL"
			} else {
				url, err = c.direct.TryLogin(ctx, goal.token, goal.flags)
				f = "TryLogin"
			}
			if err != nil {
				health.SetAuthRoutineInError(err)
				report(err, f)
				bo.BackOff(ctx, err)
				continue
			}
			if url != "" {
				// goal.url ought to be empty here.
				// However, not all control servers get this right,
				// and logging about it here just generates noise.
				c.mu.Lock()
				c.loginGoal = &LoginGoal{
					wantLoggedIn: true,
					flags:        LoginDefault,
					url:          url,
				}
				c.state = StateURLVisitRequired
				c.synced = false
				c.mu.Unlock()

				c.sendStatus("authRoutine-url", err, url, nil)
				if goal.url == url {
					// The server sent us the same URL we already tried,
					// backoff to avoid a busy loop.
					bo.BackOff(ctx, errors.New("login URL not changing"))
				} else {
					bo.BackOff(ctx, nil)
				}
				continue
			}

			// success
			health.SetAuthRoutineInError(nil)
			c.mu.Lock()
			c.loggedIn = true
			c.loginGoal = nil
			c.state = StateAuthenticated
			c.mu.Unlock()

			c.sendStatus("authRoutine-success", nil, "", nil)
			c.restartMap()
			bo.BackOff(ctx, nil)
		}
	}
}

// Expiry returns the credential expiration time, or the zero time if
// the expiration time isn't known. Used in tests only.
func (c *Auto) Expiry() *time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.expiry
}

// Direct returns the underlying direct client object. Used in tests
// only.
func (c *Auto) Direct() *Direct {
	return c.direct
}

// unpausedChanLocked returns a new channel that is closed when the
// current Auto pause is unpaused.
//
// c.mu must be held
func (c *Auto) unpausedChanLocked() <-chan struct{} {
	unpaused := make(chan struct{})
	c.unpauseWaiters = append(c.unpauseWaiters, unpaused)
	return unpaused
}

// mapRoutineState is the state of Auto.mapRoutine while it's running.
type mapRoutineState struct {
	c  *Auto
	bo *backoff.Backoff
}

func (mrs mapRoutineState) UpdateFullNetmap(nm *netmap.NetworkMap) {
	c := mrs.c
	health.SetInPollNetMap(true)

	c.mu.Lock()
	ctx := c.mapCtx
	c.synced = true
	if c.loggedIn {
		c.state = StateSynchronized
	}
	c.expiry = ptr.To(nm.Expiry)
	stillAuthed := c.loggedIn
	c.logf("[v1] mapRoutine: netmap received: %s", c.state)
	c.mu.Unlock()

	if stillAuthed {
		c.sendStatus("mapRoutine-got-netmap", nil, "", nm)
	}
	// Reset the backoff timer if we got a netmap.
	mrs.bo.BackOff(ctx, nil)
}

// mapRoutine is responsible for keeping a read-only streaming connection to the
// control server, and keeping the netmap up to date.
func (c *Auto) mapRoutine() {
	defer close(c.mapDone)
	mrs := &mapRoutineState{
		c:  c,
		bo: backoff.NewBackoff("mapRoutine", c.logf, 30*time.Second),
	}

	for {
		if err := c.waitUnpause("mapRoutine"); err != nil {
			c.logf("mapRoutine: exiting")
			return
		}

		c.mu.Lock()
		c.logf("[v1] mapRoutine: %s", c.state)
		loggedIn := c.loggedIn
		ctx := c.mapCtx
		c.mu.Unlock()

		select {
		case <-c.quit:
			c.logf("mapRoutine: quit")
			return
		default:
		}

		report := func(err error, msg string) {
			c.logf("[v1] %s: %v", msg, err)
			err = fmt.Errorf("%s: %w", msg, err)
			// don't send status updates for context errors,
			// since context cancelation is always on purpose.
			if ctx.Err() == nil {
				c.sendStatus("mapRoutine1", err, "", nil)
			}
		}

		if !loggedIn {
			// Wait for something interesting to happen
			c.mu.Lock()
			c.synced = false
			// c.state is set by authRoutine()
			c.mu.Unlock()

			select {
			case <-ctx.Done():
				c.logf("[v1] mapRoutine: context done.")
			case <-c.newMapCh:
				c.logf("[v1] mapRoutine: new map needed while idle.")
			}
		} else {
			health.SetInPollNetMap(false)

			err := c.direct.PollNetMap(ctx, mrs)

			health.SetInPollNetMap(false)
			c.mu.Lock()
			c.synced = false
			if c.state == StateSynchronized {
				c.state = StateAuthenticated
			}
			paused := c.paused
			c.mu.Unlock()

			if paused {
				mrs.bo.BackOff(ctx, nil)
				c.logf("mapRoutine: paused")
				continue
			}

			report(err, "PollNetMap")
			mrs.bo.BackOff(ctx, err)
			continue
		}
	}
}

func (c *Auto) AuthCantContinue() bool {
	if c == nil {
		return true
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	return !c.loggedIn && (c.loginGoal == nil || c.loginGoal.url != "")
}

func (c *Auto) SetHostinfo(hi *tailcfg.Hostinfo) {
	if hi == nil {
		panic("nil Hostinfo")
	}
	if !c.direct.SetHostinfo(hi) {
		// No changes. Don't log.
		return
	}

	// Send new Hostinfo to server
	c.updateControl()
}

func (c *Auto) SetNetInfo(ni *tailcfg.NetInfo) {
	if ni == nil {
		panic("nil NetInfo")
	}
	if !c.direct.SetNetInfo(ni) {
		return
	}

	// Send new NetInfo to server
	c.updateControl()
}

// SetTKAHead updates the TKA head hash that map-request infrastructure sends.
func (c *Auto) SetTKAHead(headHash string) {
	if !c.direct.SetTKAHead(headHash) {
		return
	}

	// Send new TKAHead to server
	c.updateControl()
}

func (c *Auto) sendStatus(who string, err error, url string, nm *netmap.NetworkMap) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	state := c.state
	loggedIn := c.loggedIn
	synced := c.synced
	c.inSendStatus++
	c.mu.Unlock()

	c.logf("[v1] sendStatus: %s: %v", who, state)

	var p *persist.PersistView
	var loginFin, logoutFin *empty.Message
	if state == StateAuthenticated {
		loginFin = new(empty.Message)
	}
	if state == StateNotAuthenticated {
		logoutFin = new(empty.Message)
	}
	if nm != nil && loggedIn && synced {
		p = ptr.To(c.direct.GetPersist())
	} else {
		// don't send netmap status, as it's misleading when we're
		// not logged in.
		nm = nil
	}
	new := Status{
		LoginFinished:  loginFin,
		LogoutFinished: logoutFin,
		URL:            url,
		Persist:        p,
		NetMap:         nm,
		State:          state,
		Err:            err,
	}
	c.statusFunc(new)

	c.mu.Lock()
	c.inSendStatus--
	c.mu.Unlock()
}

func (c *Auto) Login(t *tailcfg.Oauth2Token, flags LoginFlags) {
	c.logf("client.Login(%v, %v)", t != nil, flags)

	c.mu.Lock()
	c.loginGoal = &LoginGoal{
		wantLoggedIn: true,
		token:        t,
		flags:        flags,
	}
	c.mu.Unlock()

	c.cancelAuth()
}

func (c *Auto) StartLogout() {
	c.logf("client.StartLogout()")

	c.mu.Lock()
	c.loginGoal = &LoginGoal{
		wantLoggedIn: false,
	}
	c.mu.Unlock()
	c.cancelAuth()
}

func (c *Auto) Logout(ctx context.Context) error {
	c.logf("client.Logout()")

	errc := make(chan error, 1)

	c.mu.Lock()
	c.loginGoal = &LoginGoal{
		wantLoggedIn:    false,
		loggedOutResult: errc,
	}
	c.mu.Unlock()
	c.cancelAuth()

	timer, timerChannel := c.clock.NewTimer(10 * time.Second)
	defer timer.Stop()
	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
		return ctx.Err()
	case <-timerChannel:
		return context.DeadlineExceeded
	}
}

func (c *Auto) SetExpirySooner(ctx context.Context, expiry time.Time) error {
	return c.direct.SetExpirySooner(ctx, expiry)
}

// UpdateEndpoints sets the client's discovered endpoints and sends
// them to the control server if they've changed.
//
// It does not retain the provided slice.
func (c *Auto) UpdateEndpoints(endpoints []tailcfg.Endpoint) {
	changed := c.direct.SetEndpoints(endpoints)
	if changed {
		c.updateControl()
	}
}

func (c *Auto) Shutdown() {
	c.logf("client.Shutdown()")

	c.mu.Lock()
	inSendStatus := c.inSendStatus
	closed := c.closed
	direct := c.direct
	if !closed {
		c.closed = true
	}
	c.mu.Unlock()

	c.logf("client.Shutdown: inSendStatus=%v", inSendStatus)
	if !closed {
		c.unregisterHealthWatch()
		close(c.quit)
		c.cancelAuth()
		<-c.authDone
		c.cancelMap()
		<-c.mapDone
		<-c.updateDone
		if direct != nil {
			direct.Close()
		}
		c.logf("Client.Shutdown done.")
	}
}

// NodePublicKey returns the node public key currently in use. This is
// used exclusively in tests.
func (c *Auto) TestOnlyNodePublicKey() key.NodePublic {
	priv := c.direct.GetPersist()
	return priv.PrivateNodeKey().Public()
}

func (c *Auto) TestOnlySetAuthKey(authkey string) {
	c.direct.mu.Lock()
	defer c.direct.mu.Unlock()
	c.direct.authKey = authkey
}

func (c *Auto) TestOnlyTimeNow() time.Time {
	return c.clock.Now()
}

// SetDNS sends the SetDNSRequest request to the control plane server,
// requesting a DNS record be created or updated.
func (c *Auto) SetDNS(ctx context.Context, req *tailcfg.SetDNSRequest) error {
	return c.direct.SetDNS(ctx, req)
}

func (c *Auto) DoNoiseRequest(req *http.Request) (*http.Response, error) {
	return c.direct.DoNoiseRequest(req)
}

// GetSingleUseNoiseRoundTripper returns a RoundTripper that can be only be used
// once (and must be used once) to make a single HTTP request over the noise
// channel to the coordination server.
//
// In addition to the RoundTripper, it returns the HTTP/2 channel's early noise
// payload, if any.
func (c *Auto) GetSingleUseNoiseRoundTripper(ctx context.Context) (http.RoundTripper, *tailcfg.EarlyNoise, error) {
	return c.direct.GetSingleUseNoiseRoundTripper(ctx)
}
