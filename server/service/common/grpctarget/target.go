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

package grpctarget

import (
	"fmt"
	"os"
	"strings"
)

// NormalizeWorkerTarget validates a plaintext gRPC worker_target and applies optional host rewrites.
// Rejects HTTP(S) URLs. Accepts host:port and other native gRPC dial targets (e.g. dns:///...).
func NormalizeWorkerTarget(target string) (string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return "", fmt.Errorf("worker_target is empty")
	}
	lower := strings.ToLower(target)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return "", fmt.Errorf("HTTP(S) worker_target %q rejected; use a plaintext gRPC target (host:port)", target)
	}

	autofixHost := os.Getenv("AUTO_FIX_WORKER_URL")
	if autofixHost != "" {
		target = strings.Replace(target, "localhost", autofixHost, 1)
		target = strings.Replace(target, "127.0.0.1", autofixHost, 1)
	}
	autofixPortEnv := os.Getenv("AUTO_FIX_WORKER_PORT_FROM_ENV")
	if autofixPortEnv != "" {
		envVal := os.Getenv(autofixPortEnv)
		target = strings.Replace(target, "$"+autofixPortEnv+"$", envVal, 1)
	}
	return target, nil
}
