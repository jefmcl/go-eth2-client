// Copyright © 2020, 2021 Attestant Limited.
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

package phase0_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jefmcl/go-eth2-client/spec/phase0"
	"github.com/goccy/go-yaml"
	"github.com/golang/snappy"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func TestCheckpointJSON(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		err   string
	}{
		{
			name: "Empty",
			err:  "unexpected end of JSON input",
		},
		{
			name:  "JSONBad",
			input: []byte("[]"),
			err:   "invalid JSON: json: cannot unmarshal array into Go value of type phase0.checkpointJSON",
		},
		{
			name:  "EpochMissing",
			input: []byte(`{"root":"0x000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"}`),
			err:   "epoch missing",
		},
		{
			name:  "EpochWrongType",
			input: []byte(`{"epoch":true,"root":"0x000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"}`),
			err:   "invalid JSON: json: cannot unmarshal bool into Go struct field checkpointJSON.epoch of type string",
		},
		{
			name:  "EpochInvalid",
			input: []byte(`{"epoch":"-1","root":"0x000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"}`),
			err:   "invalid value for epoch: strconv.ParseUint: parsing \"-1\": invalid syntax",
		},
		{
			name:  "RootMissing",
			input: []byte(`{"epoch":"1"}`),
			err:   "root missing",
		},
		{
			name:  "RootWrongType",
			input: []byte(`{"epoch":"1","root":true}`),
			err:   "invalid JSON: json: cannot unmarshal bool into Go struct field checkpointJSON.root of type string",
		},
		{
			name:  "RootInvalid",
			input: []byte(`{"epoch":"1","root":"invalid"}`),
			err:   "invalid value for root: encoding/hex: invalid byte: U+0069 'i'",
		},
		{
			name:  "RootShort",
			input: []byte(`{"epoch":"1","root":"0x0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"}`),
			err:   "incorrect length for root",
		},
		{
			name:  "RootLong",
			input: []byte(`{"epoch":"1","root":"0x00000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"}`),
			err:   "incorrect length for root",
		},
		{
			name:  "Good",
			input: []byte(`{"epoch":"1","root":"0x000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"}`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var res phase0.Checkpoint
			err := json.Unmarshal(test.input, &res)
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
				rt, err := json.Marshal(&res)
				require.NoError(t, err)
				assert.Equal(t, string(test.input), string(rt))
			}
		})
	}
}

func TestCheckpointYAML(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		root  []byte
		err   string
	}{
		{
			name:  "Good",
			input: []byte(`{epoch: 1, root: '0x000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f'}`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var res phase0.Checkpoint
			err := yaml.Unmarshal(test.input, &res)
			if test.err != "" {
				require.EqualError(t, err, test.err)
			} else {
				require.NoError(t, err)
				rt, err := yaml.Marshal(&res)
				require.NoError(t, err)
				assert.Equal(t, string(rt), res.String())
				rt = bytes.TrimSuffix(rt, []byte("\n"))
				assert.Equal(t, string(test.input), string(rt))
			}
		})
	}
}

func TestCheckpointSpec(t *testing.T) {
	if os.Getenv("ETH2_SPEC_TESTS_DIR") == "" {
		t.Skip("ETH2_SPEC_TESTS_DIR not suppplied, not running spec tests")
	}
	baseDir := filepath.Join(os.Getenv("ETH2_SPEC_TESTS_DIR"), "tests", "mainnet", "phase0", "ssz_static", "Checkpoint", "ssz_random")
	require.NoError(t, filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if path == baseDir {
			// Only interested in subdirectories.
			return nil
		}
		require.NoError(t, err)
		if info.IsDir() {
			t.Run(info.Name(), func(t *testing.T) {
				specYAML, err := os.ReadFile(filepath.Join(path, "value.yaml"))
				require.NoError(t, err)
				var res phase0.Checkpoint
				require.NoError(t, yaml.Unmarshal(specYAML, &res))

				compressedSpecSSZ, err := os.ReadFile(filepath.Join(path, "serialized.ssz_snappy"))
				require.NoError(t, err)
				var specSSZ []byte
				specSSZ, err = snappy.Decode(specSSZ, compressedSpecSSZ)
				require.NoError(t, err)

				unmarshalled := &phase0.Checkpoint{}
				require.NoError(t, unmarshalled.UnmarshalSSZ(specSSZ))
				remarshalled, err := unmarshalled.MarshalSSZ()
				require.NoError(t, err)
				require.Equal(t, specSSZ, remarshalled)

				ssz, err := res.MarshalSSZ()
				require.NoError(t, err)
				require.Equal(t, specSSZ, ssz)

				root, err := res.HashTreeRoot()
				require.NoError(t, err)
				rootsYAML, err := os.ReadFile(filepath.Join(path, "roots.yaml"))
				require.NoError(t, err)
				require.Equal(t, string(rootsYAML), fmt.Sprintf("{root: '%#x'}\n", root))
			})
		}
		return nil
	}))
}
