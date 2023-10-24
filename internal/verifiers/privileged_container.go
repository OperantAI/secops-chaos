/*
Copyright © 2023 Operant AI

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package verifiers

import (
	"context"

	"k8s.io/client-go/kubernetes"
)

type PrivilegedContainer struct{}

func (p *PrivilegedContainer) Name() string {
	return "PrivilegedContainer"
}

func (p *PrivilegedContainer) Verify(ctx context.Context, client *kubernetes.Clientset) error {
	return nil
}

var _ Verifier = (*PrivilegedContainer)(nil)
