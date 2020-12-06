// Copyright 2017 The Mellium Contributors.
// Use of this source code is governed by the BSD 2-clause
// license that can be found in the LICENSE file.

package form

import (
	"encoding/xml"

	"mellium.im/xmlstream"
)

// A field represents a data field that may be added to a form.
type field struct {
	XMLName  xml.Name   `xml:"field"`
	Typ      string     `xml:"type,attr"`
	Var      string     `xml:"var,attr,omitempty"`
	Label    string     `xml:"label,attr,omitempty"`
	Desc     string     `xml:"desc,omitempty"`
	Value    []string   `xml:"value,omitempty"`
	Required *struct{}  `xml:"required,omitempty"`
	Field    []fieldopt `xml:"option,omitempty"`
}

func (f *field) TokenReader() xml.TokenReader {
	attr := []xml.Attr{{
		Name:  xml.Name{Local: "type"},
		Value: f.Typ,
	}}
	if f.Var != "" {
		attr = append(attr, xml.Attr{
			Name:  xml.Name{Local: "var"},
			Value: f.Var,
		})
	}
	if f.Label != "" {
		attr = append(attr, xml.Attr{
			Name:  xml.Name{Local: "label"},
			Value: f.Label,
		})
	}
	if f.Desc != "" {
		attr = append(attr, xml.Attr{
			Name:  xml.Name{Local: "desc"},
			Value: f.Desc,
		})
	}
	var child []xml.TokenReader
	for _, val := range f.Value {
		if val == "" {
			continue
		}
		child = append(child, xmlstream.Wrap(
			xmlstream.Token(xml.CharData(val)),
			xml.StartElement{Name: xml.Name{Local: "value"}},
		))
	}
	if f.Required != nil {
		child = append(child, xmlstream.Wrap(
			nil,
			xml.StartElement{Name: xml.Name{Local: "required"}},
		))
	}
	for _, opt := range f.Field {
		child = append(child, xmlstream.Wrap(
			xmlstream.Token(xml.CharData(opt.Value)),
			xml.StartElement{Name: xml.Name{Space: NS, Local: "option"}},
		))
	}

	return xmlstream.Wrap(
		xmlstream.MultiReader(child...),
		xml.StartElement{
			Name: xml.Name{Local: "field"},
			Attr: attr,
		},
	)
}

type fieldopt struct {
	XMLName xml.Name `xml:"jabber:x:data option"`
	Value   string   `xml:"value,omitempty"`
}

func newField(typ, id string, o ...Option) func(data *Data) {
	return func(data *Data) {
		f := field{
			Typ: typ,
			Var: id,
		}
		getFieldOpts(&f, o...)
		data.children = append(data.children, &f)
	}
}

// Boolean fields enable an entity to gather or provide an either-or choice
// between two options.
func Boolean(id string, o ...Option) Field {
	return newField("boolean", "", o...)
}

// Fixed is intended for data description (e.g., human-readable text such as
// "section" headers) rather than data gathering or provision.
func Fixed(o ...Option) Field {
	return newField("fixed", "", o...)
}

// Hidden fields are not shown by the form-submitting entity, but instead are
// returned, generally unmodified, with the form.
func Hidden(id string, o ...Option) Field {
	return newField("hidden", id, o...)
}

// JIDMulti enables an entity to gather or provide multiple Jabber IDs.
func JIDMulti(id string, o ...Option) Field {
	return newField("jid-multi", id, o...)
}

// JID enables an entity to gather or provide a Jabber ID.
func JID(id string, o ...Option) Field {
	return newField("jid-single", id, o...)
}

// ListMulti enables an entity to gather or provide one or more entries from a
// list.
func ListMulti(id string, o ...Option) Field {
	return newField("list-multi", id, o...)
}

// ListSingle enables an entity to gather or provide a single entry from a list.
func ListSingle(id string, o ...Option) Field {
	return newField("list-single", id, o...)
}

// TextMulti enables an entity to gather or provide multiple lines of text.
func TextMulti(id string, o ...Option) Field {
	return newField("text-multi", id, o...)
}

// TextPrivate enables an entity to gather or provide a line of text that should
// be obscured in the submitting entities interface (eg. with multiple
// asterisks).
func TextPrivate(id string, o ...Option) Field {
	return newField("text-private", id, o...)
}

// TextSingle enables an entity to gather or provide a line of text.
func TextSingle(id string, o ...Option) Field {
	return newField("text-single", id, o...)
}
