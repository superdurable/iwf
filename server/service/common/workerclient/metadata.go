// Copyright (c) 2022-2026 Super Durable, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package workerclient

import (
	"fmt"
	"strings"
	"unicode"
)

// ValidateDefaultHeaders rejects metadata keys that gRPC cannot send.
func ValidateDefaultHeaders(headers map[string]string) error {
	for key := range headers {
		if err := validateMetadataKey(key); err != nil {
			return err
		}
	}
	return nil
}

func validateMetadataKey(key string) error {
	if key == "" {
		return fmt.Errorf("defaultHeaders: empty metadata key")
	}
	lower := strings.ToLower(key)
	if strings.HasPrefix(lower, "grpc-") {
		return fmt.Errorf("defaultHeaders: key %q must not start with grpc-", key)
	}
	for _, r := range key {
		if unicode.IsUpper(r) {
			return fmt.Errorf("defaultHeaders: key %q must be lowercase", key)
		}
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' && r != '-' && r != '.' {
			return fmt.Errorf("defaultHeaders: key %q has invalid character %q", key, r)
		}
	}
	return nil
}
