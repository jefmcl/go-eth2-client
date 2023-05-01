// Copyright © 2023 Attestant Limited.
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

package deneb

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/attestantio/go-eth2-client/spec/altair"
	"github.com/attestantio/go-eth2-client/spec/capella"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/pkg/errors"
)

// beaconBlockBodyJSON is the spec representation of the struct.
type beaconBlockBodyJSON struct {
	RANDAOReveal          string                                `json:"randao_reveal"`
	ETH1Data              *phase0.ETH1Data                      `json:"eth1_data"`
	Graffiti              string                                `json:"graffiti"`
	ProposerSlashings     []*phase0.ProposerSlashing            `json:"proposer_slashings"`
	AttesterSlashings     []*phase0.AttesterSlashing            `json:"attester_slashings"`
	Attestations          []*phase0.Attestation                 `json:"attestations"`
	Deposits              []*phase0.Deposit                     `json:"deposits"`
	VoluntaryExits        []*phase0.SignedVoluntaryExit         `json:"voluntary_exits"`
	SyncAggregate         *altair.SyncAggregate                 `json:"sync_aggregate"`
	ExecutionPayload      *ExecutionPayload                     `json:"execution_payload"`
	BLSToExecutionChanges []*capella.SignedBLSToExecutionChange `json:"bls_to_execution_changes"`
	BlobKzgCommitments    []string                              `json:"blob_kzg_commitments"`
}

// MarshalJSON implements json.Marshaler.
func (b *BeaconBlockBody) MarshalJSON() ([]byte, error) {
	blobKzgCommitments := make([]string, len(b.BlobKzgCommitments))
	for i := range b.BlobKzgCommitments {
		blobKzgCommitments[i] = b.BlobKzgCommitments[i].String()
	}

	return json.Marshal(&beaconBlockBodyJSON{
		RANDAOReveal:          fmt.Sprintf("%#x", b.RANDAOReveal),
		ETH1Data:              b.ETH1Data,
		Graffiti:              fmt.Sprintf("%#x", b.Graffiti),
		ProposerSlashings:     b.ProposerSlashings,
		AttesterSlashings:     b.AttesterSlashings,
		Attestations:          b.Attestations,
		Deposits:              b.Deposits,
		VoluntaryExits:        b.VoluntaryExits,
		SyncAggregate:         b.SyncAggregate,
		ExecutionPayload:      b.ExecutionPayload,
		BLSToExecutionChanges: b.BLSToExecutionChanges,
		BlobKzgCommitments:    blobKzgCommitments,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *BeaconBlockBody) UnmarshalJSON(input []byte) error {
	var data beaconBlockBodyJSON
	if err := json.Unmarshal(input, &data); err != nil {
		return errors.Wrap(err, "invalid JSON")
	}
	return b.unpack(&data)
}

func (b *BeaconBlockBody) unpack(data *beaconBlockBodyJSON) error {
	if data.RANDAOReveal == "" {
		return errors.New("RANDAO reveal missing")
	}
	randaoReveal, err := hex.DecodeString(strings.TrimPrefix(data.RANDAOReveal, "0x"))
	if err != nil {
		return errors.Wrap(err, "invalid value for RANDAO reveal")
	}
	if len(randaoReveal) != phase0.SignatureLength {
		return errors.New("incorrect length for RANDAO reveal")
	}
	copy(b.RANDAOReveal[:], randaoReveal)
	if data.ETH1Data == nil {
		return errors.New("ETH1 data missing")
	}
	b.ETH1Data = data.ETH1Data
	if data.Graffiti == "" {
		return errors.New("graffiti missing")
	}
	graffiti, err := hex.DecodeString(strings.TrimPrefix(data.Graffiti, "0x"))
	if err != nil {
		return errors.Wrap(err, "invalid value for graffiti")
	}
	if len(graffiti) != phase0.GraffitiLength {
		return errors.New("incorrect length for graffiti")
	}
	copy(b.Graffiti[:], graffiti)
	if data.ProposerSlashings == nil {
		return errors.New("proposer slashings missing")
	}
	b.ProposerSlashings = data.ProposerSlashings
	if data.AttesterSlashings == nil {
		return errors.New("attester slashings missing")
	}
	b.AttesterSlashings = data.AttesterSlashings
	if data.Attestations == nil {
		return errors.New("attestations missing")
	}
	b.Attestations = data.Attestations
	if data.Deposits == nil {
		return errors.New("deposits missing")
	}
	b.Deposits = data.Deposits
	if data.VoluntaryExits == nil {
		return errors.New("voluntary exits missing")
	}
	b.VoluntaryExits = data.VoluntaryExits
	if data.SyncAggregate == nil {
		return errors.New("sync aggregate missing")
	}
	b.SyncAggregate = data.SyncAggregate
	if data.ExecutionPayload == nil {
		return errors.New("execution payload missing")
	}
	b.ExecutionPayload = data.ExecutionPayload
	b.BLSToExecutionChanges = data.BLSToExecutionChanges
	if data.BlobKzgCommitments == nil {
		return errors.New("blob kzg commitments missing")
	}
	b.BlobKzgCommitments = make([]KzgCommitment, len(data.BlobKzgCommitments))
	for i := range data.BlobKzgCommitments {
		data, err := hex.DecodeString(strings.TrimPrefix(data.BlobKzgCommitments[i], "0x"))
		if err != nil {
			return errors.Wrap(err, "failed to parse blob KZG commitment")
		}
		if len(data) != KzgCommitmentLength {
			return errors.New("incorrect length for blob KZG commitment")
		}
		copy(b.BlobKzgCommitments[i][:], data)
	}

	return nil
}