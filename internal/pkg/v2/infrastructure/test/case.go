/*******************************************************************************
 * Copyright 2020 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package test

import (
	"net/http/httptest"
	"testing"
)

// PostCondition defines the signature of the post-condition function provided to the test case.
type PostCondition func(t *testing.T, w *httptest.ResponseRecorder)

// Case contains references to dependencies required by the test case.
type Case struct {
	name           string
	request        []byte
	postCondition  PostCondition
	expectedStatus int
}

// New is a factory function that returns a Case struct.
func New(name string, request []byte, postCondition PostCondition, expectedStatus int) *Case {
	return &Case{
		name:           name,
		request:        request,
		postCondition:  postCondition,
		expectedStatus: expectedStatus,
	}
}

// Name returns the test case's name.
func (c *Case) Name() string {
	return c.name
}

// Request returns the test case's request.
func (c *Case) Request() []byte {
	return c.request
}

// PostCondition executes the test case's post-condition function.
func (c *Case) PostCondition(t *testing.T, w *httptest.ResponseRecorder) {
	c.postCondition(t, w)
}

// ExpectedStatus returns the test case's expected status.
func (c *Case) ExpectedStatus() int {
	return c.expectedStatus
}
