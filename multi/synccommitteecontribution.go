// Copyright © 2021 Attestant Limited.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package multi

import (
	"context"

	consensusclient "github.com/jefmcl/go-eth2-client"
	"github.com/jefmcl/go-eth2-client/spec/altair"
	"github.com/jefmcl/go-eth2-client/spec/phase0"
)

// SyncCommitteeContribution provides a sync committee contribution.
func (s *Service) SyncCommitteeContribution(ctx context.Context,
	slot phase0.Slot,
	subcommitteeIndex uint64,
	beaconBlockRoot phase0.Root,
) (
	*altair.SyncCommitteeContribution,
	error,
) {
	res, err := s.doCall(ctx, func(ctx context.Context, client consensusclient.Service) (interface{}, error) {
		block, err := client.(consensusclient.SyncCommitteeContributionProvider).SyncCommitteeContribution(ctx, slot, subcommitteeIndex, beaconBlockRoot)
		if err != nil {
			return nil, err
		}
		return block, nil
	}, nil)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}
	return res.(*altair.SyncCommitteeContribution), nil
}
