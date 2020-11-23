// Copyright 2019 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package marshal_test

import (
	"encoding/xml"
	"testing"

	"mellium.im/xmpp/internal/marshal"
	"mellium.im/xmpp/stanza"
)

type testWriteFlusher struct {
	flushes int
}

func (wf *testWriteFlusher) EncodeToken(t xml.Token) error {
	return nil
}

func (wf *testWriteFlusher) Flush() error {
	wf.flushes++
	return nil
}

func TestFlushes(t *testing.T) {
	t.Run("EncodeXML", func(t *testing.T) {
		f := &testWriteFlusher{}
		if err := marshal.EncodeXML(f, 1); err != nil {
			t.Fatal(err)
		}

		if f.flushes != 1 {
			t.Errorf("Expected 1 flush call got %d", f.flushes)
		}
	})
	t.Run("EncodeXMLElement", func(t *testing.T) {
		f := &testWriteFlusher{}
		start := xml.StartElement{Name: xml.Name{Local: "int"}}
		if err := marshal.EncodeXMLElement(f, 1, start); err != nil {
			t.Fatal(err)
		}

		if f.flushes != 1 {
			t.Errorf("Expected 1 flush call got %d", f.flushes)
		}
	})
}

func TestMarshalTokenReader(t *testing.T) {
	// If the payload to marshal is already a TokenReader, just return it.
	r := stanza.IQ{}.Wrap(nil)
	rr, err := marshal.TokenReader(r)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if r != rr {
		t.Errorf("got different xml.TokenReader out: want=%v, got=%v", r, rr)
	}
}
