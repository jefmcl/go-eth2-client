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
	"bytes"
	"fmt"

	"github.com/jefmcl/go-eth2-client/spec/capella"
	"github.com/goccy/go-yaml"
)

// executionPayloadYAML is the spec representation of the struct.
type executionPayloadYAML struct {
	ParentHash    string                `yaml:"parent_hash"`
	FeeRecipient  string                `yaml:"fee_recipient"`
	StateRoot     string                `yaml:"state_root"`
	ReceiptsRoot  string                `yaml:"receipts_root"`
	LogsBloom     string                `yaml:"logs_bloom"`
	PrevRandao    string                `yaml:"prev_randao"`
	BlockNumber   uint64                `yaml:"block_number"`
	GasLimit      uint64                `yaml:"gas_limit"`
	GasUsed       uint64                `yaml:"gas_used"`
	Timestamp     uint64                `yaml:"timestamp"`
	ExtraData     string                `yaml:"extra_data"`
	BaseFeePerGas string                `yaml:"base_fee_per_gas"`
	BlockHash     string                `yaml:"block_hash"`
	Transactions  []string              `yaml:"transactions"`
	Withdrawals   []*capella.Withdrawal `yaml:"withdrawals"`
	ExcessDataGas string                `yaml:"excess_data_gas"`
}

// MarshalYAML implements yaml.Marshaler.
func (e *ExecutionPayload) MarshalYAML() ([]byte, error) {
	transactions := make([]string, len(e.Transactions))
	for i := range e.Transactions {
		transactions[i] = fmt.Sprintf("%#x", e.Transactions[i])
	}

	extraData := "0x"
	if len(e.ExtraData) > 0 {
		extraData = fmt.Sprintf("%#x", e.ExtraData)
	}

	yamlBytes, err := yaml.MarshalWithOptions(&executionPayloadYAML{
		ParentHash:    e.ParentHash.String(),
		FeeRecipient:  e.FeeRecipient.String(),
		StateRoot:     e.StateRoot.String(),
		ReceiptsRoot:  e.ReceiptsRoot.String(),
		LogsBloom:     fmt.Sprintf("%#x", e.LogsBloom),
		PrevRandao:    fmt.Sprintf("%#x", e.PrevRandao),
		BlockNumber:   e.BlockNumber,
		GasLimit:      e.GasLimit,
		GasUsed:       e.GasUsed,
		Timestamp:     e.Timestamp,
		ExtraData:     extraData,
		BaseFeePerGas: e.BaseFeePerGas.Dec(),
		BlockHash:     fmt.Sprintf("%#x", e.BlockHash),
		Transactions:  transactions,
		Withdrawals:   e.Withdrawals,
		ExcessDataGas: e.ExcessDataGas.Dec(),
	}, yaml.Flow(true))
	if err != nil {
		return nil, err
	}
	return bytes.ReplaceAll(yamlBytes, []byte(`"`), []byte(`'`)), nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (e *ExecutionPayload) UnmarshalYAML(input []byte) error {
	// We unmarshal to the JSON struct to save on duplicate code.
	var data executionPayloadJSON
	if err := yaml.Unmarshal(input, &data); err != nil {
		return err
	}
	return e.unpack(&data)
}
