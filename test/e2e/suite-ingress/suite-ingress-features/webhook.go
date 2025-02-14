// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ingress

import (
	"fmt"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"github.com/apache/apisix-ingress-controller/test/e2e/scaffold"
)

var _ = ginkgo.Describe("suite-ingress-features: Enable webhooks", func() {
	s := scaffold.NewScaffold(&scaffold.Options{
		Name:                  "webhook",
		IngressAPISIXReplicas: 1,
		ApisixResourceVersion: scaffold.ApisixResourceVersion().V2,
		EnableWebhooks:        true,
	})

	ginkgo.It("should fail to create the ApisixRoute with invalid plugin configuration", func() {
		backendSvc, backendPorts := s.DefaultHTTPBackend()
		ar := fmt.Sprintf(`
apiVersion: apisix.apache.org/v2
kind: ApisixRoute
metadata:
 name: httpbin-route
spec:
 http:
 - name: rule1
   match:
     hosts:
     - httpbin.org
     paths:
       - /status/*
   backends:
   - serviceName: %s
     servicePort: %d
   plugins:
   - name: api-breaker
     enable: true
     config:
       break_response_code: 1000 # should in [200, 599]
`, backendSvc, backendPorts[0])

		err := s.CreateResourceFromString(ar)
		assert.Error(ginkgo.GinkgoT(), err, "Failed to create ApisixRoute")
		assert.Contains(ginkgo.GinkgoT(), err.Error(), "denied the request")
		assert.Contains(ginkgo.GinkgoT(), err.Error(), "api-breaker plugin's config is invalid")
		assert.Contains(ginkgo.GinkgoT(), err.Error(), "Must be less than or equal to 599")
	})

	ginkgo.It("should fail to update the ApisixRoute with invalid plugin configuration", func() {
		backendSvc, backendPorts := s.DefaultHTTPBackend()
		ar := fmt.Sprintf(`
apiVersion: apisix.apache.org/v2
kind: ApisixRoute
metadata:
 name: httpbin-route
spec:
 http:
 - name: rule1
   match:
     hosts:
     - httpbin.org
     paths:
       - /status/*
   backends:
   - serviceName: %s
     servicePort: %d
   plugins:
   - name: echo
     enable: true
     config:
       body: "successsful"
`, backendSvc, backendPorts[0])
		assert.Nil(ginkgo.GinkgoT(), s.CreateResourceFromString(ar))
		assert.Nil(ginkgo.GinkgoT(), s.EnsureNumApisixRoutesCreated(1), "ApisixRoute should be 1")

		ar = fmt.Sprintf(`
apiVersion: apisix.apache.org/v2
kind: ApisixRoute
metadata:
 name: httpbin-route
spec:
 http:
 - name: rule1
   match:
     hosts:
     - httpbin.org
     paths:
       - /status/*
   backends:
   - serviceName: %s
     servicePort: %d
   plugins:
   - name: echo
     enable: true
     config:
       body_info: "failed"
`, backendSvc, backendPorts[0])
		err := s.CreateResourceFromString(ar)
		assert.Error(ginkgo.GinkgoT(), err, "Failed to update ApisixRoute")
		assert.Contains(ginkgo.GinkgoT(), err.Error(), "denied the request")
		assert.Contains(ginkgo.GinkgoT(), err.Error(), "echo plugin's config is invalid")
	})

	ginkgo.It("should fail to create the ApisixPluginConfig with invalid plugin configuration", func() {
		apc := `
apiVersion: apisix.apache.org/v2
kind: ApisixPluginConfig
metadata:
  name: echo
spec:
  plugins:
  - name: echo
    enable: true
    config:
      body-failed: "failed"
`

		err := s.CreateResourceFromString(apc)
		assert.Error(ginkgo.GinkgoT(), err, "Failed to create ApisixRoute")
		assert.Contains(ginkgo.GinkgoT(), err.Error(), "denied the request")
		assert.Contains(ginkgo.GinkgoT(), err.Error(), "echo plugin's config is invalid")
	})

	ginkgo.It("should fail to update the ApisixPluginConfig with invalid plugin configuration", func() {
		apc := `
apiVersion: apisix.apache.org/v2
kind: ApisixPluginConfig
metadata:
  name: echo
spec:
  plugins:
  - name: echo
    enable: true
    config:
      before_body: "This is the preface"
      after_body: "This is the epilogue"
      headers:
        X-Foo: v1
        X-Foo2: v2
  - name: cors
    enable: true
`
		assert.Nil(ginkgo.GinkgoT(), s.CreateResourceFromString(apc), "creatint a ApisixPluginConfig")
		assert.Nil(ginkgo.GinkgoT(), s.EnsureNumApisixPluginConfigCreated(1), "ApisixPluginConfig should be 1")

		apc = `
apiVersion: apisix.apache.org/v2
kind: ApisixPluginConfig
metadata:
  name: echo
spec:
  plugins:
  - name: echo
    enable: true
    config:
      body-failed: "failed"
  - name: cors
    enable: true
`
		err := s.CreateResourceFromString(apc)
		assert.Error(ginkgo.GinkgoT(), err, "Failed to create ApisixRoute")
		assert.Contains(ginkgo.GinkgoT(), err.Error(), "denied the request")
		assert.Contains(ginkgo.GinkgoT(), err.Error(), "echo plugin's config is invalid")
	})
})
