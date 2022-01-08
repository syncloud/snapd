// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2014-2015 Canonical Ltd
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

package osutil_test

import (
	"github.com/snapcore/snapd/asserts"
	. "gopkg.in/check.v1"
)

type FileDigestSuite struct{}

var _ = Suite(&FileDigestSuite{})

func (ts *FileDigestSuite) TestFileDigest(c *C) {
	//exData := []byte("hashmeplease")

	//tempdir := c.MkDir()
	//fn := filepath.Join(tempdir, "ex.file")
	//err := ioutil.WriteFile(fn, exData, 0644)
	//c.Assert(err, IsNil)

	//digest, _, err := osutil.FileDigest("/home/boris/files_135_amd64.snap", crypto.SHA3_384)
	sha3_384, size, err := asserts.SnapFileSHA3_384("/home/boris/files_135_amd64.snap")
	c.Assert(err, IsNil)
	c.Assert(size, IsNil)
	//c.Check(size, Equals, uint64(len(exData)))
	//h512 := sha512.Sum512(exData)
	c.Assert(sha3_384, NotNil)
}
