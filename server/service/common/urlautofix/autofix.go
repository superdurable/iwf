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

package urlautofix

import (
	"os"
	"strings"
)

type FixWorkerUrlFunc func(url string) string

var workerUrlFixer FixWorkerUrlFunc = DefaultFixWorkerUrlFunc

func SetWorkerUrlFixer(fixer FixWorkerUrlFunc) {
	workerUrlFixer = fixer
}

func FixWorkerUrl(url string) string {
	return workerUrlFixer(url)
}

func DefaultFixWorkerUrlFunc(url string) string {
	autofixUrl := os.Getenv("AUTO_FIX_WORKER_URL")
	if autofixUrl != "" {
		url = strings.Replace(url, "localhost", autofixUrl, 1)
		url = strings.Replace(url, "127.0.0.1", autofixUrl, 1)
	}
	autofixPortEnv := os.Getenv("AUTO_FIX_WORKER_PORT_FROM_ENV")
	if autofixPortEnv != "" {
		envVal := os.Getenv(autofixPortEnv)
		url = strings.Replace(url, "$"+autofixPortEnv+"$", envVal, 1)
	}
	url = strings.TrimRight(url, "/")

	return url
}
