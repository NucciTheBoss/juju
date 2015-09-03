// Copyright 2014 Canonical Ltd. All rights reserved.
// Licensed under the AGPLv3, see LICENCE file for details.

package apiserver_test

import (
	"github.com/juju/names"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/juju/apiserver/authentication"
	"github.com/juju/juju/apiserver/params"
	"github.com/juju/juju/juju/testing"
	"github.com/juju/juju/state"
	"github.com/juju/juju/testing/factory"
)

type agentAuthenticatorSuite struct {
	testing.JujuConnSuite
}
type userFinder struct {
	user state.Entity
}

func (u userFinder) FindEntity(tag names.Tag) (state.Entity, error) {
	return u.user, nil
}

var _ = gc.Suite(&agentAuthenticatorSuite{})

func (s *agentAuthenticatorSuite) TestAuthenticatorForTag(c *gc.C) {
	fact := factory.NewFactory(s.State)
	user := fact.MakeUser(c, &factory.UserParams{Password: "password"})
	srv := newServer(c, s.State)
	defer srv.Stop()
	authenticator, err := srv.AuthenticatorForTag(user.Tag())
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(authenticator, gc.NotNil)
	userFinder := userFinder{user}

	entity, err := authenticator.Authenticate(userFinder, user.Tag(), params.LoginRequest{
		Credentials: "password",
		Nonce:       "nonce",
	})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(entity, gc.DeepEquals, user)
}

func (s *agentAuthenticatorSuite) TestAuthenticatorForTagGetsMacaroonAuthenticator(c *gc.C) {
	srv := newServer(c, s.State)
	defer srv.Stop()
	authenticator, err := srv.AuthenticatorForTag(nil)
	c.Assert(err, jc.ErrorIsNil)
	_, ok := authenticator.(*authentication.MacaroonAuthenticator)
	c.Assert(ok, jc.IsTrue)
}

func (s *agentAuthenticatorSuite) TestMachineGetsAgentAuthentictor(c *gc.C) {
	srv := newServer(c, s.State)
	defer srv.Stop()
	authenticator, err := srv.AuthenticatorForTag(names.NewMachineTag("0"))
	c.Assert(err, jc.ErrorIsNil)
	_, ok := authenticator.(*authentication.AgentAuthenticator)
	c.Assert(ok, jc.IsTrue)
}
func (s *agentAuthenticatorSuite) TestUnitGetsAgentAuthentictor(c *gc.C) {
	srv := newServer(c, s.State)
	defer srv.Stop()
	authenticator, err := srv.AuthenticatorForTag(names.NewUnitTag("wordpress/0"))
	c.Assert(err, jc.ErrorIsNil)
	_, ok := authenticator.(*authentication.AgentAuthenticator)
	c.Assert(ok, jc.IsTrue)
}
func (s *agentAuthenticatorSuite) TestNotSupportedTag(c *gc.C) {
	srv := newServer(c, s.State)
	defer srv.Stop()
	authenticator, err := srv.AuthenticatorForTag(names.NewServiceTag("not-support"))
	c.Assert(err, gc.ErrorMatches, "invalid request")
	c.Assert(authenticator, gc.IsNil)
}
