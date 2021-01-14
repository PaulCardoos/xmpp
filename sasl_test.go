// Copyright 2016 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package xmpp_test

import (
	"bytes"
	"context"
	"encoding/xml"
	"strings"
	"testing"

	"mellium.im/sasl"
	"mellium.im/xmpp"
	"mellium.im/xmpp/internal/ns"
	"mellium.im/xmpp/internal/saslerr"
)

func TestSASLPanicsNoMechanisms(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected call to SASL() with no mechanisms to panic")
		}
	}()
	_ = xmpp.SASL("", "")
}

func TestSASLList(t *testing.T) {
	b := &bytes.Buffer{}
	e := xml.NewEncoder(b)
	start := xml.StartElement{Name: xml.Name{Space: ns.SASL, Local: "mechanisms"}}
	s := xmpp.SASL("", "", sasl.Plain, sasl.ScramSha256)
	req, err := s.List(context.Background(), e, start)
	switch {
	case err != nil:
		t.Fatal(err)
	case req != true:
		t.Error("Expected SASL to be a required feature")
	}
	if err = e.Flush(); err != nil {
		t.Fatal(err)
	}

	// Mechanisms should be printed exactly thus:
	if !bytes.Contains(b.Bytes(), []byte(`<mechanism>PLAIN</mechanism>`)) {
		t.Error("Expected mechanisms list to include PLAIN")
	}
	if !bytes.Contains(b.Bytes(), []byte(`<mechanism>SCRAM-SHA-256</mechanism>`)) {
		t.Error("Expected mechanisms list to include SCRAM-SHA-256")
	}

	// The wrapper can be a bit more flexible as long as the mechanisms are there.
	d := xml.NewDecoder(b)
	tok, err := d.Token()
	if err != nil {
		t.Fatal(err)
	}
	se := tok.(xml.StartElement)
	if se.Name.Local != "mechanisms" || se.Name.Space != ns.SASL {
		t.Errorf("Unexpected name for mechanisms start element: %+v", se.Name)
	}
	// Skip two mechanisms
	_, err = d.Token()
	if err != nil {
		t.Fatal(err)
	}
	err = d.Skip()
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.Token()
	if err != nil {
		t.Fatal(err)
	}
	err = d.Skip()
	if err != nil {
		t.Fatal(err)
	}

	// Check the end token.
	tok, err = d.Token()
	if err != nil {
		t.Fatal(err)
	}
	_ = tok.(xml.EndElement)
}

func TestSASLParse(t *testing.T) {
	s := xmpp.SASL("", "", sasl.Plain)
	for _, test := range []struct {
		xml   string
		items []string
		err   bool
	}{
		{`<mechanisms xmlns='urn:ietf:params:xml:ns:xmpp-sasl'>
		<mechanism>EXTERNAL</mechanism>
		<mechanism>SCRAM-SHA-1-PLUS</mechanism>
		<mechanism>SCRAM-SHA-1</mechanism>
		<mechanism>PLAIN</mechanism>
		</mechanisms>`, []string{"EXTERNAL", "PLAIN", "SCRAM-SHA-1-PLUS", "SCRAM-SHA-1"}, false},
		{`<oops xmlns='urn:ietf:params:xml:ns:xmpp-sasl'><mechanism>PLAIN</mechanism></oop>`, nil, true},
		{`<mechanisms xmlns='badns'><mechanism>PLAIN</mechanism></mechanisms>`, nil, true},
		{`<mechanisms xmlns='urn:ietf:params:xml:ns:xmpp-sasl'><mechanism xmlns="nope">PLAIN</mechanism></mechanisms>`, []string{}, false},
	} {
		r := strings.NewReader(test.xml)
		d := xml.NewDecoder(r)
		tok, _ := d.Token()
		start := tok.(xml.StartElement)
		req, list, err := s.Parse(context.Background(), d, &start)
		switch {
		case test.err && err == nil:
			t.Error("Expected sasl mechanism parsing to error")
		case !test.err && err != nil:
			t.Error(err)
		case req != true:
			t.Error("Expected parsed SASL feature to be required")
		case len(list.([]string)) != len(test.items):
			t.Errorf("Expected data to contain 4 items, got %d", len(list.([]string)))
		}
		for _, m := range test.items {
			matched := false
			for _, m2 := range list.([]string) {
				if m == m2 {
					matched = true
					break
				}
			}
			if !matched {
				t.Fatalf("Expected data to contain %v", m)
			}
		}
	}
}

func panicPerms(*sasl.Negotiator) bool {
	panic("permissions function should not be called")
}

var saslTestCases = [...]featureTestCase{
	// Simple client tests with plain.
	0: {
		sf:  xmpp.SASL("", "", sasl.Plain),
		in:  `<abb/>`,
		out: `<auth xmlns="urn:ietf:params:xml:ns:xmpp-sasl" mechanism="PLAIN">AHRlc3QA</auth>`,
		err: xmpp.ErrUnexpectedPayload,
	},
	1: {
		sf:  xmpp.SASL("", "", sasl.Plain),
		in:  `<!-- not a start element -->`,
		out: `<auth xmlns="urn:ietf:params:xml:ns:xmpp-sasl" mechanism="PLAIN">AHRlc3QA</auth>`,
		err: xmpp.ErrUnexpectedPayload,
	},
	2: {
		sf:  xmpp.SASL("", "", sasl.Plain),
		in:  `<failure xmlns="urn:ietf:params:xml:ns:xmpp-sasl"><not-authorized/></failure>`,
		out: `<auth xmlns="urn:ietf:params:xml:ns:xmpp-sasl" mechanism="PLAIN">AHRlc3QA</auth>`,
		err: saslerr.Failure{Condition: saslerr.NotAuthorized},
	},
	3: {
		sf:         xmpp.SASL("", "", sasl.Plain),
		in:         `<success xmlns="urn:ietf:params:xml:ns:xmpp-sasl"/>`,
		out:        `<auth xmlns="urn:ietf:params:xml:ns:xmpp-sasl" mechanism="PLAIN">AHRlc3QA</auth>`,
		finalState: xmpp.Authn,
		err:        saslerr.Failure{Condition: saslerr.NotAuthorized},
	},

	// Simple server tests with plain.
	4: {
		state: xmpp.Received,
		sf:    xmpp.SASLServer(panicPerms, sasl.Plain),
		in:    `<abb/>`,
		out:   `<failure xmlns="urn:ietf:params:xml:ns:xmpp-sasl"><malformed-request></malformed-request></failure>`,
		err:   xmpp.ErrUnexpectedPayload,
	},
	5: {
		state: xmpp.Received,
		sf:    xmpp.SASLServer(panicPerms, sasl.Plain),
		in:    `<!-- not a start element -->`,
		err:   xmpp.ErrUnexpectedPayload,
	},
	6: {
		// TODO: can the client send failure?
		state: xmpp.Received,
		sf:    xmpp.SASLServer(panicPerms, sasl.Plain),
		in:    `<failure xmlns="urn:ietf:params:xml:ns:xmpp-sasl"><not-authorized/></failure>`,
		err:   saslerr.Failure{Condition: saslerr.NotAuthorized},
	},
	7: {
		state: xmpp.Received,
		sf:    xmpp.SASLServer(panicPerms, sasl.Plain),
		in:    `<abort xmlns="urn:ietf:params:xml:ns:xmpp-sasl"><not-authorized/></abort>`,
		out:   `<failure xmlns="urn:ietf:params:xml:ns:xmpp-sasl"><aborted></aborted></failure>`,
		err:   xmpp.ErrTerminated,
	},
	8: {
		state: xmpp.Received,
		sf: xmpp.SASLServer(func(*sasl.Negotiator) bool {
			return false
		}, sasl.Plain),
		in:  `<auth xmlns="urn:ietf:params:xml:ns:xmpp-sasl" mechanism="PLAIN">AHRlc3QA</auth>`,
		out: `<failure xmlns="urn:ietf:params:xml:ns:xmpp-sasl"><not-authorized></not-authorized></failure>`,
		err: sasl.ErrAuthn,
	},
	9: {
		state: xmpp.Received,
		sf: xmpp.SASLServer(func(*sasl.Negotiator) bool {
			return true
		}, sasl.Plain),
		in:         `<auth xmlns="urn:ietf:params:xml:ns:xmpp-sasl" mechanism="PLAIN">AHRlc3QA</auth>`,
		out:        `<success xmlns="urn:ietf:params:xml:ns:xmpp-sasl"></success>`,
		finalState: xmpp.Authn,
	},
}

func TestSASL(t *testing.T) {
	runFeatureTests(t, saslTestCases[:])
}
