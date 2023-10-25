package convert

import (
	"encoding/json"
	"errors"
	"fmt"
)

type POETerm struct {
	Term       string            `json:"term"`
	TermPlural string            `json:"term_plural"`
	Definition POETermDefinition `json:"definition"`
}

type POETermDefinition struct {
	IsPlural bool

	Value  *string
	Plural *POETermPluralDefinition
}

func (d *POETermDefinition) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	switch v := v.(type) {
	case string:
		d.Value = &v
		return nil
	case map[string]interface{}:
		d.IsPlural = true
		return json.Unmarshal(data, &d.Plural)
	case nil:
		empty := ""
		d.Value = &empty
		return nil
	}

	return errors.New("invalid definition type")
}

func (d POETermDefinition) MarshalJSON() ([]byte, error) {
	if d.IsPlural {
		return json.Marshal(d.Plural)
	}

	return json.Marshal(d.Value)
}

type POETermPluralDefinition struct {
	Zero  *string `json:"zero,omitempty"`
	One   *string `json:"one,omitempty"`
	Two   *string `json:"two,omitempty"`
	Few   *string `json:"few,omitempty"`
	Many  *string `json:"many,omitempty"`
	Other string  `json:"other"`
}

func (p POETermPluralDefinition) Map(mapper func(string) (string, error)) (*POETermPluralDefinition, error) {
	var zero, one, two, few, many *string

	if p.Zero != nil {
		v, err := mapper(*p.Zero)
		zero = &v
		if err != nil {
			return nil, err
		}
	}
	if p.One != nil {
		v, err := mapper(*p.One)
		one = &v
		if err != nil {
			return nil, err
		}
	}
	if p.Two != nil {
		v, err := mapper(*p.Two)
		two = &v
		if err != nil {
			return nil, err
		}
	}
	if p.Few != nil {
		v, err := mapper(*p.Few)
		few = &v
		if err != nil {
			return nil, err
		}
	}
	if p.Many != nil {
		v, err := mapper(*p.Many)
		many = &v
		if err != nil {
			return nil, err
		}
	}

	v, err := mapper(p.Other)
	if err != nil {
		return nil, err
	}

	return &POETermPluralDefinition{
		Zero: zero, One: one,
		Two: two, Few: few,
		Many: many, Other: v,
	}, nil
}

func (p POETermPluralDefinition) ToICUMessageFormat() string {
	str := "{count, plural,"
	if p.Zero != nil {
		str += fmt.Sprintf(" =0 {%s}", *p.Zero)
	}
	if p.One != nil {
		str += fmt.Sprintf(" =1 {%s}", *p.One)
	}
	if p.Two != nil {
		str += fmt.Sprintf(" =2 {%s}", *p.Two)
	}
	if p.Few != nil {
		str += fmt.Sprintf(" few {%s}", *p.Few)
	}
	if p.Many != nil {
		str += fmt.Sprintf(" many {%s}", *p.Many)
	}
	str += fmt.Sprintf(" other {%s}", p.Other)
	str += "}"

	return str
}
