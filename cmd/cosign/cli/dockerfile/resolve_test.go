// Copyright 2021 The Sigstore Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dockerfile

import (
	"bytes"
	"reflect"
	"testing"
)

func Test_resolveDigest(t *testing.T) {
	tests := []struct {
		name       string
		dockerfile string
		want       string
		wantErr    bool
	}{
		{
			"happy alpine",
			`FROM alpine:3.13`,
			`FROM index.docker.io/library/alpine@sha256:469b6e04ee185740477efa44ed5bdd64a07bbdd6c7e5f5d169e540889597b911
`,
			false,
		},
		{
			"happy alpine trim",
			`   FROM    alpine:3.13   `,
			`FROM    index.docker.io/library/alpine@sha256:469b6e04ee185740477efa44ed5bdd64a07bbdd6c7e5f5d169e540889597b911
`,
			false,
		},
		{
			"happy alpine copy",
			`FROM alpine:3.13
COPY --from=alpine:3.13
`,
			`FROM index.docker.io/library/alpine@sha256:469b6e04ee185740477efa44ed5bdd64a07bbdd6c7e5f5d169e540889597b911
COPY --from=index.docker.io/library/alpine@sha256:469b6e04ee185740477efa44ed5bdd64a07bbdd6c7e5f5d169e540889597b911
`,
			false,
		},
		{
			"alpine with digest",
			`FROM alpine@sha256:469b6e04ee185740477efa44ed5bdd64a07bbdd6c7e5f5d169e540889597b911`,
			`FROM alpine@sha256:469b6e04ee185740477efa44ed5bdd64a07bbdd6c7e5f5d169e540889597b911
`,
			false,
		},
		{
			"multi-line",
			`FROM alpine:3.13
COPY . .

RUN ls`,
			`FROM index.docker.io/library/alpine@sha256:469b6e04ee185740477efa44ed5bdd64a07bbdd6c7e5f5d169e540889597b911
COPY . .

RUN ls
`,
			false,
		},
		{
			"skip scratch",
			`FROM alpine:3.13
FROM scratch
RUN ls`,
			`FROM index.docker.io/library/alpine@sha256:469b6e04ee185740477efa44ed5bdd64a07bbdd6c7e5f5d169e540889597b911
FROM scratch
RUN ls
`,
			false,
		},
		{
			"should not break invalid image ref",
			`FROM alpine:$(TAG)
FROM $(IMAGE)
`,
			`FROM alpine:$(TAG)
FROM $(IMAGE)
`,
			false,
		},
		{
			"should not break for invalid --from image reference",
			`FROM golang:latest AS builder
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /go/src/github.com/foo/bar/app .
CMD ["./app"]`,
			`FROM index.docker.io/library/golang@sha256:660f138b4477001d65324a51fa158c1b868651b44e43f0953bf062e9f38b72f3 AS builder
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM index.docker.io/library/alpine@sha256:8914eb54f968791faf6a8638949e480fef81e697984fba772b3976835194c6d4
WORKDIR /root/
COPY --from=builder /go/src/github.com/foo/bar/app .
CMD ["./app"]
`,
			false,
		},
		{
			"should not break for invalid --from image reference with digest",
			`COPY --from=nginx:latest /etc/nginx/nginx.conf /nginx.conf`,
			`COPY --from=index.docker.io/library/nginx@sha256:0047b729188a15da49380d9506d65959cce6d40291ccfb4e039f5dc7efd33286 /etc/nginx/nginx.conf /nginx.conf
`,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveDigest(bytes.NewBuffer([]byte(tt.dockerfile)))
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveDigest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(string(got), tt.want) {
				t.Errorf("resolveDigest() got = %v, want %v", string(got), tt.want)
			}
		})
	}
}
