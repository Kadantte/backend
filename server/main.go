// Copyright (C) 2024 Pavel Sobolev
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package server

import (
	"morbo/errors"
	"morbo/log"
)

func Main(args []string) error {
	server, err := NewServer("0.0.0.0", 80)
	if err != nil {
		log.Error.Println("failed to create the server")
		return errors.Error
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Error.Println("failed to listen and serve")
		return errors.Error
	}

	return nil
}
