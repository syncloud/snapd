// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2014-2016 Canonical Ltd
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

package syncloud

import (
	"encoding/json"
)

// a stringList is something that can be deserialized from a JSON
// []string or a string, like the values of the "extra" documents in
// error responses
type stringList []string

func (sish *stringList) UnmarshalJSON(bs []byte) error {
	var ss []string
	e1 := json.Unmarshal(bs, &ss)
	if e1 == nil {
		*sish = stringList(ss)
		return nil
	}

	var s string
	e2 := json.Unmarshal(bs, &s)
	if e2 == nil {
		*sish = stringList([]string{s})
		return nil
	}

	return e1
}

type ssoMsg struct {
	Code    string                `json:"code"`
	Message string                `json:"message"`
	Extra   map[string]stringList `json:"extra"`
}
