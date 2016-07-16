//    Gonews is a webapp that provides a forum where users can post and discuss links
//
//    Copyright (C) 2016  mparaiso <mparaiso@online.fr>

//    This program is free software: you can redistribute it and/or modify
//    it under the terms of the GNU Affero General Public License as published
//    by the Free Software Foundation, either version 3 of the License, or
//    (at your option) any later version.

//    This program is distributed in the hope that it will be useful,
//    but WITHOUT ANY WARRANTY; without even the implied warranty of
//    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//    GNU Affero General Public License for more details.

//    You should have received a copy of the GNU Affero General Public License
//    along with this program.  If not, see <http://www.gnu.org/licenses/>.

package gonews

// CSRFGenerator generates and validate csrf tokens
type CSRFGenerator interface {
	Generate(userID, actionID string) string
	Valid(token, userID, actionID string) bool
}

// UserFinder can find users from a datasource
type UserFinder interface {
	GetOneByEmail(string) (*User, error)
	GetOneByUsername(string) (*User, error)
}
