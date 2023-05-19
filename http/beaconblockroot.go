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

package http

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jefmcl/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
)

type beaconBlockRootJSON struct {
	Data *beaconBlockRootDataJSON `json:"data"`
}

type beaconBlockRootDataJSON struct {
	Root string `json:"root"`
}

// BeaconBlockRoot fetches a block's root given a block ID.
// N.B if a signed beacon block for the block ID is not available this will return nil without an error.
func (s *Service) BeaconBlockRoot(ctx context.Context, blockID string) (*phase0.Root, error) {
	respBodyReader, err := s.get(ctx, fmt.Sprintf("/eth/v1/beacon/blocks/%s/root", blockID))
	if err != nil {
		return nil, errors.Wrap(err, "failed to request beacon block root")
	}
	if respBodyReader == nil {
		return nil, nil
	}

	var beaconBlockRootJSON beaconBlockRootJSON
	if err := json.NewDecoder(respBodyReader).Decode(&beaconBlockRootJSON); err != nil {
		return nil, errors.Wrap(err, "failed to parse beacon block root")
	}

	if beaconBlockRootJSON.Data == nil {
		return nil, errors.New("no data returned")
	}
	if beaconBlockRootJSON.Data.Root == "" {
		return nil, errors.New("no root returned")
	}

	bytes, err := hex.DecodeString(strings.TrimPrefix(beaconBlockRootJSON.Data.Root, "0x"))
	if err != nil {
		return nil, errors.Wrap(err, "invalid root returned")
	}

	var res phase0.Root
	copy(res[:], bytes)

	return &res, nil
}
