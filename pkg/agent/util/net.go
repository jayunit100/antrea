// Copyright 2019 Antrea Authors
//
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

package util

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
)

const (
	interfaceNameLength   = 15
	interfacePrefixLength = 8
	interfaceKeyLen       = interfaceNameLength - (interfacePrefixLength + 1)
)

func generateInterfaceName(id string, prefix string) string {
	hash := sha1.New()
	io.WriteString(hash, id)
	interfaceKey := hex.EncodeToString(hash.Sum(nil))
	if len(prefix) > interfacePrefixLength {
		prefix = prefix[:interfacePrefixLength]
	}
	return fmt.Sprintf("%s-%s", prefix, interfaceKey[:interfaceKeyLen])
}

// GenerateContainerInterfaceName generates a unique interface name using the
// Pod's Namespace and name. The output should be deterministic (so that
// multiple calls to GenerateContainerInterfaceName with the same parameters
// return the same value). The output has the length of interfaceNameLength(15).
// The probability of collision should be neglectable.
func GenerateContainerInterfaceName(podName string, podNamespace string) string {
	id := fmt.Sprintf("pod/%s/%s", podNamespace, podName)
	return generateInterfaceName(id, podName)
}

// GenerateTunnelInterfaceName generates a unique interface name for the tunnel
// to the Node, using the Node's name.
func GenerateTunnelInterfaceName(nodeName string) string {
	id := fmt.Sprintf("node/%s", nodeName)
	return generateInterfaceName(id, nodeName)
}
