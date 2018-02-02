/*
   Copyright (c) 2015 Andrey Sibiryov <me@kobology.ru>
   Copyright (c) 2015 Other contributors as noted in the AUTHORS file.

   This file is part of GORB - Go Routing and Balancing.

   GORB is free software; you can redistribute it and/or modify
   it under the terms of the GNU Lesser General Public License as published by
   the Free Software Foundation; either version 3 of the License, or
   (at your option) any later version.

   GORB is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
   GNU Lesser General Public License for more details.

   You should have received a copy of the GNU Lesser General Public License
   along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"encoding/json"
	"net/http"

	"github.com/kobolog/gorb/core"
	"github.com/kobolog/gorb/util"

	"errors"
	"net"
	"strings"

	"github.com/containous/traefik/log"
	"github.com/gorilla/mux"
	"github.com/kobolog/gorb/pulse"
	"github.com/kobolog/gorb/types"
)

var badRequest = errors.New("malformed request")

type errorResponse struct {
	Error string `json:"error"`
}

type serviceRequest struct {
	Host      string `json:"host"`
	Port      uint16 `json:"port"`
	Protocol  string `json:"protocol"`
	Scheduler string `json:"scheduler"`
	Flags     string `json:"flags"`
}

type backendRequest struct {
	Host         string         `json:"host"`
	Port         uint16         `json:"port"`
	Weight       uint32         `json:"weight"`
	Method       string         `json:"method"`
	PulseOptions *pulse.Options `json:"pulse"`
}

func writeJSON(w http.ResponseWriter, obj interface{}) {
	w.Header().Add("Content-Type", "application/json")
	w.Write(util.MustMarshal(obj, util.JSONOptions{Indent: true}))
}

func writeError(w http.ResponseWriter, err error) {
	var code int

	switch err {
	case badRequest:
		code = http.StatusBadRequest
	case core.ErrObjectExists:
		code = http.StatusConflict
	case core.ErrObjectNotFound:
		code = http.StatusNotFound
	default:
		code = http.StatusInternalServerError
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(util.MustMarshal(&errorResponse{err.Error()}, util.JSONOptions{Indent: true}))
}

func fillInService(svc *types.Service, vars map[string]string, req serviceRequest) error {
	svc.StoreID = vars["vsID"]
	if len(svc.StoreID) == 0 {
		log.Warnf("vsID is required, but was empty")
		return badRequest
	}
	vip, err := net.LookupIP(req.Host)
	if err != nil {
		log.Warnf("invalid host %q: %v", req.Host, err)
		return badRequest
	}
	svc.VIP = vip[0]
	svc.Port = req.Port
	svc.Protocol = req.Protocol
	svc.Scheduler = req.Scheduler
	svc.Flags = strings.Split(req.Flags, "|")
	return nil
}

type serviceCreateHandler struct {
	ctx *core.Context
}

func (h serviceCreateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		req  serviceRequest
		svc  types.Service
		vars = mux.Vars(r)
	)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, err)
	}

	if err := fillInService(&svc, vars, req); err != nil {
		writeError(w, err)
	}

	if err := h.ctx.CreateService(&svc); err != nil {
		writeError(w, err)
	}
}

type serviceUpdateHandler struct {
	ctx *core.Context
}

func (h serviceUpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		req  serviceRequest
		svc  types.Service
		vars = mux.Vars(r)
	)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, err)
	}

	if err := fillInService(&svc, vars, req); err != nil {
		writeError(w, err)
	}

	if err := h.ctx.UpdateService(&svc); err != nil {
		writeError(w, err)
	}
}

type backendCreateHandler struct {
	ctx *core.Context
}

func fillInBackend(backend *types.Backend, vars map[string]string, req backendRequest) error {
	backend.StoreID = vars["rsID"]
	if len(backend.StoreID) == 0 {
		log.Warnf("rsID is required, but was empty")
		return badRequest
	}
	ips, err := net.LookupIP(req.Host)
	if err != nil {
		log.Warnf("invalid host %q: %v", req.Host, err)
		return badRequest
	}
	backend.IP = ips[0]
	backend.Port = req.Port
	backend.Weight = req.Weight
	backend.Forward = req.Method
	backend.PulseOptions = req.PulseOptions
	return nil
}

func (h backendCreateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		req     backendRequest
		backend types.Backend
		vars    = mux.Vars(r)
	)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, err)
	}

	if err := fillInBackend(&backend, vars, req); err != nil {
		writeError(w, err)
	}

	if err := h.ctx.CreateBackend(vars["vsID"], &backend); err != nil {
		writeError(w, err)
	}
}

type backendUpdateHandler struct {
	ctx *core.Context
}

func (h backendUpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		req     backendRequest
		backend types.Backend
		vars    = mux.Vars(r)
	)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, err)
	}

	if err := fillInBackend(&backend, vars, req); err != nil {
		writeError(w, err)
	}

	if err := h.ctx.UpdateBackend(vars["vsID"], &backend); err != nil {
		writeError(w, err)
	}
}

type serviceRemoveHandler struct {
	ctx *core.Context
}

func (h serviceRemoveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if _, err := h.ctx.RemoveService(vars["vsID"]); err != nil {
		writeError(w, err)
	}
}

type backendRemoveHandler struct {
	ctx *core.Context
}

func (h backendRemoveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if _, err := h.ctx.RemoveBackend(vars["vsID"], vars["rsID"]); err != nil {
		writeError(w, err)
	}
}

type serviceListHandler struct {
	ctx *core.Context
}

func (h serviceListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if list, err := h.ctx.ListServices(); err != nil {
		writeError(w, err)
	} else {
		writeJSON(w, list)
	}
}

type serviceStatusHandler struct {
	ctx *core.Context
}

func (h serviceStatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if opts, err := h.ctx.GetService(vars["vsID"]); err != nil {
		writeError(w, err)
	} else {
		writeJSON(w, opts)
	}
}

type backendStatusHandler struct {
	ctx *core.Context
}

func (h backendStatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if opts, err := h.ctx.GetBackend(vars["vsID"], vars["rsID"]); err != nil {
		writeError(w, err)
	} else {
		writeJSON(w, opts)
	}
}
