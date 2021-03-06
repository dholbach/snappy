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

package snapstate

import (
	"github.com/ubuntu-core/snappy/progress"
	"github.com/ubuntu-core/snappy/snap"
	"github.com/ubuntu-core/snappy/snappy"
	"github.com/ubuntu-core/snappy/store"
)

type managerBackend interface {
	// install releated
	Download(name, channel string, checker func(*snap.Info) error, meter progress.Meter, auther store.Authenticator) (*snap.Info, string, error)
	CheckSnap(snapFilePath string, curInfo *snap.Info, flags int) error
	SetupSnap(snapFilePath string, si *snap.SideInfo, flags int) error
	CopySnapData(newSnap, oldSnap *snap.Info, flags int) error
	LinkSnap(info *snap.Info) error
	// the undoers for install
	UndoSetupSnap(s snap.PlaceInfo) error
	UndoCopySnapData(newSnap *snap.Info, flags int) error

	// remove releated
	CanRemove(info *snap.Info, active bool) bool
	UnlinkSnap(info *snap.Info, meter progress.Meter) error
	RemoveSnapFiles(s snap.PlaceInfo, meter progress.Meter) error
	RemoveSnapData(info *snap.Info) error

	// testing helpers
	Candidate(sideInfo *snap.SideInfo)
}

type defaultBackend struct{}

func (b *defaultBackend) Candidate(*snap.SideInfo) {}

func (b *defaultBackend) Download(name, channel string, checker func(*snap.Info) error, meter progress.Meter, auther store.Authenticator) (*snap.Info, string, error) {
	mStore := snappy.NewConfiguredUbuntuStoreSnapRepository()
	snap, err := mStore.Snap(name, channel, auther)
	if err != nil {
		return nil, "", err
	}

	err = checker(snap)
	if err != nil {
		return nil, "", err
	}

	downloadedSnapFile, err := mStore.Download(snap, meter, auther)
	if err != nil {
		return nil, "", err
	}

	return snap, downloadedSnapFile, nil
}

func (b *defaultBackend) CheckSnap(snapFilePath string, curInfo *snap.Info, flags int) error {
	meter := &progress.NullProgress{}
	return snappy.CheckSnap(snapFilePath, curInfo, snappy.InstallFlags(flags), meter)
}

func (b *defaultBackend) SetupSnap(snapFilePath string, sideInfo *snap.SideInfo, flags int) error {
	meter := &progress.NullProgress{}
	_, err := snappy.SetupSnap(snapFilePath, sideInfo, snappy.InstallFlags(flags), meter)
	return err
}

func (b *defaultBackend) CopySnapData(newInfo, oldInfo *snap.Info, flags int) error {
	meter := &progress.NullProgress{}
	return snappy.CopyData(newInfo, oldInfo, snappy.InstallFlags(flags), meter)
}

func (b *defaultBackend) LinkSnap(info *snap.Info) error {
	meter := &progress.NullProgress{}
	return snappy.LinkSnap(info, meter)
}

func (b *defaultBackend) UndoSetupSnap(s snap.PlaceInfo) error {
	meter := &progress.NullProgress{}
	snappy.UndoSetupSnap(s, meter)
	return nil
}

func (b *defaultBackend) UndoCopySnapData(newInfo *snap.Info, flags int) error {
	meter := &progress.NullProgress{}
	snappy.UndoCopyData(newInfo, snappy.InstallFlags(flags), meter)
	return nil
}

func (b *defaultBackend) CanRemove(info *snap.Info, active bool) bool {
	return snappy.CanRemove(info, active)
}

func (b *defaultBackend) UnlinkSnap(info *snap.Info, meter progress.Meter) error {
	return snappy.UnlinkSnap(info, meter)
}

func (b *defaultBackend) RemoveSnapFiles(s snap.PlaceInfo, meter progress.Meter) error {
	return snappy.RemoveSnapFiles(s, meter)
}

func (b *defaultBackend) RemoveSnapData(info *snap.Info) error {
	return snappy.RemoveSnapData(info)
}
