// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package auth

import (
	"bytes"
	"fmt"
	"net/http"
	"sort"

	"github.com/ubuntu-core/snappy/overlord/state"
)

// AuthState represents current authenticated users as tracked in state
type AuthState struct {
	LastID int         `json:"last-id"`
	Users  []UserState `json:"users"`
}

// UserState represents an authenticated user
type UserState struct {
	ID              int      `json:"id"`
	Username        string   `json:"username,omitempty"`
	Macaroon        string   `json:"macaroon,omitempty"`
	Discharges      []string `json:"discharges,omitempty"`
	StoreMacaroon   string   `json:"store-macaroon,omitempty"`
	StoreDischarges []string `json:"store-discharges,omitempty"`
}

// NewUser tracks a new authenticated user and saves its details in the state
func NewUser(st *state.State, username, macaroon string, discharges []string) (*UserState, error) {
	var authStateData AuthState

	err := st.Get("auth", &authStateData)
	if err == state.ErrNoState {
		authStateData = AuthState{}
	} else if err != nil {
		return nil, err
	}

	sort.Strings(discharges)
	authStateData.LastID++
	authenticatedUser := UserState{
		ID:              authStateData.LastID,
		Username:        username,
		Macaroon:        macaroon,
		Discharges:      discharges,
		StoreMacaroon:   macaroon,
		StoreDischarges: discharges,
	}
	authStateData.Users = append(authStateData.Users, authenticatedUser)

	st.Set("auth", authStateData)

	return &authenticatedUser, nil
}

// RemoveUser removes a user from the state given its ID
func RemoveUser(st *state.State, userID int) error {
	var authStateData AuthState

	err := st.Get("auth", &authStateData)
	if err != nil {
		return err
	}

	for i := range authStateData.Users {
		if authStateData.Users[i].ID == userID {
			// delete without preserving order
			n := len(authStateData.Users) - 1
			authStateData.Users[i] = authStateData.Users[n]
			authStateData.Users[n] = UserState{}
			authStateData.Users = authStateData.Users[:n]
			st.Set("auth", authStateData)
			return nil
		}
	}

	return fmt.Errorf("invalid user")
}

// User returns a user from the state given its ID
func User(st *state.State, id int) (*UserState, error) {
	var authStateData AuthState

	err := st.Get("auth", &authStateData)
	if err != nil {
		return nil, err
	}

	for _, user := range authStateData.Users {
		if user.ID == id {
			return &user, nil
		}
	}
	return nil, fmt.Errorf("invalid user")
}

var ErrInvalidAuth = fmt.Errorf("invalid authentication")

// CheckMacaroon returns the UserState for the given macaroon/discharges credentials
func CheckMacaroon(st *state.State, macaroon string, discharges []string) (*UserState, error) {
	var authStateData AuthState
	err := st.Get("auth", &authStateData)
	if err != nil {
		return nil, ErrInvalidAuth
	}

NextUser:
	for _, user := range authStateData.Users {
		if user.Macaroon != macaroon {
			continue
		}
		if len(user.Discharges) != len(discharges) {
			continue
		}
		// sort discharges (stored users' discharges are already sorted)
		sort.Strings(discharges)
		for i, d := range user.Discharges {
			if d != discharges[i] {
				continue NextUser
			}
		}
		return &user, nil
	}
	return nil, ErrInvalidAuth
}

// Authenticator returns MacaroonAuthenticator for current authenticated user represented by UserState
func (us *UserState) Authenticator() *MacaroonAuthenticator {
	return newMacaroonAuthenticator(us.StoreMacaroon, us.StoreDischarges)
}

// MacaroonAuthenticator is a store authenticator based on macaroons
type MacaroonAuthenticator struct {
	Macaroon   string
	Discharges []string
}

func newMacaroonAuthenticator(macaroon string, discharges []string) *MacaroonAuthenticator {
	return &MacaroonAuthenticator{
		Macaroon:   macaroon,
		Discharges: discharges,
	}
}

// Authenticate will add the store expected Authorization header for macaroons
func (ma *MacaroonAuthenticator) Authenticate(r *http.Request) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, `Macaroon root="%s"`, ma.Macaroon)
	for _, discharge := range ma.Discharges {
		fmt.Fprintf(&buf, `, discharge="%s"`, discharge)
	}
	r.Header.Set("Authorization", buf.String())
}
