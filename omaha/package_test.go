// Copyright 2015 CoreOS, Inc.
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

package omaha

import (
	"strings"
	"testing"

	"encoding/xml"

	"github.com/kylelemons/godebug/pretty"
	"github.com/stretchr/testify/assert"
)

func TestPackageFromPath(t *testing.T) {
	expect := Package{
		Name:     "null",
		SHA1:     "2jmj7l5rSw0yVb/vlWAYkK/YBwk=",
		SHA256:   "47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
		Size:     0,
		Required: false,
	}

	p := Package{}
	if err := p.FromPath("/dev/null"); err != nil {
		t.Fatal(err)
	}

	if diff := pretty.Compare(expect, p); diff != "" {
		t.Errorf("Hashing /dev/null failed: %v", diff)
	}
}

func TestProtocolFromReader(t *testing.T) {
	data := strings.NewReader("testing\n")
	expect := Package{
		Name:     "",
		SHA1:     "mAFznarkTsUpPU4fU9P00tQm2Rw=",
		SHA256:   "EqYfThc/s6EcBdZHH3Ryj3YjG0pfzZZnzvOvh6OuTcI=",
		Size:     8,
		Required: false,
	}

	p := Package{}
	if err := p.FromReader(data); err != nil {
		t.Fatal(err)
	}

	if diff := pretty.Compare(expect, p); diff != "" {
		t.Errorf("Hashing failed: %v", diff)
	}
}

func TestPackageVerify(t *testing.T) {
	p := Package{
		Name:     "null",
		SHA1:     "2jmj7l5rSw0yVb/vlWAYkK/YBwk=",
		SHA256:   "47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
		Size:     0,
		Required: false,
	}

	if err := p.Verify("/dev"); err != nil {
		t.Fatal(err)
	}
}

func TestPackageVerifyNoSHA256(t *testing.T) {
	p := Package{
		Name:     "null",
		SHA1:     "2jmj7l5rSw0yVb/vlWAYkK/YBwk=",
		SHA256:   "",
		Size:     0,
		Required: false,
	}

	if err := p.Verify("/dev"); err != nil {
		t.Fatal(err)
	}
}

func TestPackageVerifyBadSize(t *testing.T) {
	p := Package{
		Name:     "null",
		SHA1:     "2jmj7l5rSw0yVb/vlWAYkK/YBwk=",
		SHA256:   "47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
		Size:     1,
		Required: false,
	}

	err := p.Verify("/dev")
	if err == nil {
		t.Error("verify passed")
	}
	if err != PackageSizeMismatchError {
		t.Error(err)
	}

}

func TestPackageVerifyBadSHA1(t *testing.T) {
	p := Package{
		Name:     "null",
		SHA1:     "xxxxxxxxxxxxxxxxxxxxxxxxxxx=",
		SHA256:   "47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
		Size:     0,
		Required: false,
	}

	err := p.Verify("/dev")
	if err == nil {
		t.Error("verify passed")
	}
	if err != PackageHashMismatchError {
		t.Error(err)
	}
}

func TestPackageVerifyBadSHA256(t *testing.T) {
	p := Package{
		Name:     "null",
		SHA1:     "2jmj7l5rSw0yVb/vlWAYkK/YBwk=",
		SHA256:   "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx=",
		Size:     0,
		Required: false,
	}

	err := p.Verify("/dev")
	if err == nil {
		t.Error("verify passed")
	}
	if err != PackageHashMismatchError {
		t.Error(err)
	}
}

func TestPackageWithMetadata(t *testing.T) {
	jsonContent := "{url: 'http://example.com'}"
	p := Package{
		Name:     "null",
		SHA1:     "2jmj7l5rSw0yVb/vlWAYkK/YBwk=",
		SHA256:   "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx=",
		Size:     0,
		Required: false,
		Metadata: &Metadata{
			ContentType: "content/json",
			Content:     jsonContent,
		},
	}

	assert.Equal(t, jsonContent, p.Metadata.Content)

	xmlPkg := []byte("<Package name=\"\" hash=\"\" hash_sha256=\"\" size=\"\" required=\"false\"><metadata content-type=\"content/json\"><![CDATA[" + jsonContent + "]]></metadata></Package>")

	var newPkg Package
	err := xml.Unmarshal(xmlPkg, &newPkg)
	assert.NoError(t, err)
	assert.Equal(t, p.Metadata.Content, newPkg.Metadata.Content)
}
