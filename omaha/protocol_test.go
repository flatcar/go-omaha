// Copyright 2013-2015 CoreOS, Inc.
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
	"encoding/xml"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

const (
	sampleRequest = `<?xml version="1.0" encoding="UTF-8"?>
<request protocol="3.0" version="ChromeOSUpdateEngine-0.1.0.0" updaterversion="ChromeOSUpdateEngine-0.1.0.0" installsource="ondemandupdate" ismachine="1">
<os version="Indy" platform="Chrome OS" sp="ForcedUpdate_x86_64"></os>
<app appid="{87efface-864d-49a5-9bb3-4b050a7c227a}" bootid="{7D52A1CC-7066-40F0-91C7-7CB6A871BFDE}" machinealias="mymachine" machineid="{8BDE4C4D-9083-4D61-B41C-3253212C0C37}" oem="ec3000" version="ForcedUpdate" track="dev-channel" from_track="developer-build" lang="en-US" board="amd64-generic" hardware_class="" delta_okay="false" >
<ping active="1" a="-1" r="-1"></ping>
<updatecheck targetversionprefix=""></updatecheck>
<event eventtype="3" eventresult="2" previousversion=""></event>
</app>
</request>
`
	sampleResponse = `<?xml version="1.0" encoding="UTF-8"?>
<response protocol="3.0">
<daystart elapsed_seconds="49008"/>
<app appid="{87efface-864d-49a5-9bb3-4b050a7c227a}" status="ok">
<ping status="ok"/>
<updatecheck status="ok">
<urls>
<url codebase="http://kam:8080/static/"/>
</urls>
<manifest version="9999.0.0">
<packages>
<package hash="+LXvjiaPkeYDLHoNKlf9qbJwvnk=" name="update.gz" size="67546213" required="true"/>
</packages>
<actions>
<action event="postinstall" DisplayVersion="9999.0.0" sha256="0VAlQW3RE99SGtSB5R4m08antAHO8XDoBMKDyxQT/Mg=" needsadmin="false" IsDeltaPayload="true" />
</actions>
</manifest>
</updatecheck>
</app>
</response>
`
)

func TestOmahaRequestUpdateCheck(t *testing.T) {
	v, err := ParseRequest("", strings.NewReader(sampleRequest))
	if err != nil {
		t.Fatalf("ParseRequest failed: %v", err)
	}

	if v.OS.Version != "Indy" {
		t.Error("Unexpected version", v.OS.Version)
	}

	if v.Apps[0].ID != "{87efface-864d-49a5-9bb3-4b050a7c227a}" {
		t.Error("Expected an App ID")
	}

	if v.Apps[0].BootID != "{7D52A1CC-7066-40F0-91C7-7CB6A871BFDE}" {
		t.Error("Expected a Boot ID")
	}

	if v.Apps[0].MachineAlias != "mymachine" {
		t.Error("Expected a Machine Alias")
	}

	if v.Apps[0].MachineID != "{8BDE4C4D-9083-4D61-B41C-3253212C0C37}" {
		t.Error("Expected a Machine ID")
	}

	if v.Apps[0].OEM != "ec3000" {
		t.Error("Expected an OEM")
	}

	if v.Apps[0].UpdateCheck == nil {
		t.Error("Expected an UpdateCheck")
	}

	if v.Apps[0].Version != "ForcedUpdate" {
		t.Error("Verison is ForcedUpdate")
	}

	if v.Apps[0].FromTrack != "developer-build" {
		t.Error("developer-build")
	}

	if v.Apps[0].Track != "dev-channel" {
		t.Error("dev-channel")
	}

	if v.Apps[0].Events[0].Type != EventTypeUpdateComplete {
		t.Error("Expected EventTypeUpdateComplete")
	}

	if v.Apps[0].Events[0].Result != EventResultSuccessReboot {
		t.Error("Expected EventResultSuccessReboot")
	}
}

func TestOmahaResponseWithUpdate(t *testing.T) {
	parsed, err := ParseResponse("", strings.NewReader(sampleResponse))
	if err != nil {
		t.Fatalf("ParseResponse failed: %v", err)
	}

	expected := &Response{
		XMLName:  xml.Name{Local: "response"},
		Protocol: "3.0",
		DayStart: DayStart{ElapsedSeconds: "49008"},
		Apps: []*AppResponse{&AppResponse{
			ID:     "{87efface-864d-49a5-9bb3-4b050a7c227a}",
			Status: AppOK,
			Ping:   &PingResponse{"ok"},
			UpdateCheck: &UpdateResponse{
				Status: UpdateOK,
				URLs: []*URL{&URL{
					CodeBase: "http://kam:8080/static/",
				}},
				Manifest: &Manifest{
					Version: "9999.0.0",
					Packages: []*Package{&Package{
						SHA1:     "+LXvjiaPkeYDLHoNKlf9qbJwvnk=",
						Name:     "update.gz",
						Size:     67546213,
						Required: true,
					}},
					Actions: []*Action{&Action{
						Event:          "postinstall",
						DisplayVersion: "9999.0.0",
						SHA256:         "0VAlQW3RE99SGtSB5R4m08antAHO8XDoBMKDyxQT/Mg=",
						IsDeltaPayload: true,
					}},
				},
			},
		}},
	}

	if !reflect.DeepEqual(parsed, expected) {
		t.Errorf("parsed != expected\n%v\n%v", parsed, expected)
	}
}

func TestOmahaResponsAsRequest(t *testing.T) {
	_, err := ParseRequest("", strings.NewReader(sampleResponse))
	if err == nil {
		t.Fatal("ParseRequest successfully parsed a response")
	}
}

func TestOmahaRequestAsResponse(t *testing.T) {
	_, err := ParseResponse("", strings.NewReader(sampleRequest))
	if err == nil {
		t.Fatal("ParseResponse successfully parsed a request")
	}
}

func ExampleNewResponse() {
	response := NewResponse()
	app := response.AddApp("{52F1B9BC-D31A-4D86-9276-CBC256AADF9A}", "ok")
	app.AddPing()
	u := app.AddUpdateCheck(UpdateOK)
	u.AddURL("http://localhost/updates")
	m := u.AddManifest("9999.0.0")
	k := m.AddPackage()
	k.SHA1 = "+LXvjiaPkeYDLHoNKlf9qbJwvnk="
	k.Name = "update.gz"
	k.Size = 67546213
	k.Required = true
	a := m.AddAction("postinstall")
	a.DisplayVersion = "9999.0.0"
	a.SHA256 = "0VAlQW3RE99SGtSB5R4m08antAHO8XDoBMKDyxQT/Mg="
	a.NeedsAdmin = false
	a.IsDeltaPayload = true
	a.DisablePayloadBackoff = true

	if raw, err := xml.MarshalIndent(response, "", " "); err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Printf("%s%s\n", xml.Header, raw)
	}

	// Output:
	// <?xml version="1.0" encoding="UTF-8"?>
	// <response protocol="3.0" server="go-omaha">
	//  <daystart elapsed_seconds="0"></daystart>
	//  <app appid="{52F1B9BC-D31A-4D86-9276-CBC256AADF9A}" status="ok">
	//   <ping status="ok"></ping>
	//   <updatecheck status="ok">
	//    <urls>
	//     <url codebase="http://localhost/updates"></url>
	//    </urls>
	//    <manifest version="9999.0.0">
	//     <packages>
	//      <package name="update.gz" hash="+LXvjiaPkeYDLHoNKlf9qbJwvnk=" size="67546213" required="true"></package>
	//     </packages>
	//     <actions>
	//      <action event="postinstall" DisplayVersion="9999.0.0" sha256="0VAlQW3RE99SGtSB5R4m08antAHO8XDoBMKDyxQT/Mg=" IsDeltaPayload="true" DisablePayloadBackoff="true"></action>
	//     </actions>
	//    </manifest>
	//   </updatecheck>
	//  </app>
	// </response>
}

func ExampleNewRequest() {
	request := NewRequest()
	request.Version = ""
	request.OS = &OS{
		Platform:    "Chrome OS",
		Version:     "Indy",
		ServicePack: "ForcedUpdate_x86_64",
	}
	app := request.AddApp("{27BD862E-8AE8-4886-A055-F7F1A6460627}", "1.0.0.0")
	app.AddUpdateCheck()

	event := app.AddEvent()
	event.Type = EventTypeDownloadComplete
	event.Result = EventResultError

	if raw, err := xml.MarshalIndent(request, "", " "); err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Printf("%s%s\n", xml.Header, raw)
	}

	// Output:
	// <?xml version="1.0" encoding="UTF-8"?>
	// <request protocol="3.0">
	//  <os platform="Chrome OS" version="Indy" sp="ForcedUpdate_x86_64"></os>
	//  <app appid="{27BD862E-8AE8-4886-A055-F7F1A6460627}" version="1.0.0.0">
	//   <updatecheck></updatecheck>
	//   <event eventtype="1" eventresult="0"></event>
	//  </app>
	// </request>
}
