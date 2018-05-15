package indiclient

import (
	"sort"
	"time"
)

// Device is an INDI device.
type Device struct {
	Name             string                    `json:"name"`
	TextProperties   map[string]TextProperty   `json:"textProperties"`
	SwitchProperties map[string]SwitchProperty `json:"switchProperties"`
	NumberProperties map[string]NumberProperty `json:"numberProperties"`
	BlobProperties   map[string]BlobProperty   `json:"blobProperties"`
	LightProperties  map[string]LightProperty  `json:"lightProperties"`
	Messages         []Message                 `json:"messages"`
}

// Message is a message received from indiserver.
type Message struct {
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
}

// TextProperty is a text property on a device.
type TextProperty struct {
	Name        string               `json:"name"`
	Label       string               `json:"label"`
	Group       string               `json:"group"`
	State       PropertyState        `json:"state"`
	Timeout     int                  `json:"timeout"`
	LastUpdated time.Time            `json:"lastUpdated"`
	Messages    []Message            `json:"messages"`
	Permissions PropertyPermission   `json:"permissions"`
	Values      map[string]TextValue `json:"values"`
}

// TextValue is a text value on a TextProperty.
type TextValue struct {
	Name  string `json:"name"`
	Label string `json:"label"`
	Value string `json:"value"`
}

// SwitchProperty is a switch property on a device.
type SwitchProperty struct {
	Name        string                 `json:"name"`
	Label       string                 `json:"label"`
	Group       string                 `json:"group"`
	State       PropertyState          `json:"state"`
	Timeout     int                    `json:"timeout"`
	LastUpdated time.Time              `json:"lastUpdated"`
	Messages    []Message              `json:"messages"`
	Rule        SwitchRule             `json:"rule"`
	Permissions PropertyPermission     `json:"permissions"`
	Values      map[string]SwitchValue `json:"values"`
}

// SwitchValue is a switch value on a SwitchProperty.
type SwitchValue struct {
	Name  string      `json:"name"`
	Label string      `json:"label"`
	Value SwitchState `json:"value"`
}

// NumberProperty is a number property on a device.
type NumberProperty struct {
	Name        string                 `json:"name"`
	Label       string                 `json:"label"`
	Group       string                 `json:"group"`
	State       PropertyState          `json:"state"`
	Timeout     int                    `json:"timeout"`
	LastUpdated time.Time              `json:"lastUpdated"`
	Messages    []Message              `json:"messages"`
	Permissions PropertyPermission     `json:"permissions"`
	Values      map[string]NumberValue `json:"values"`
}

// NumberValue is a number value on a NumberProperty.
type NumberValue struct {
	Name   string `json:"name"`
	Label  string `json:"label"`
	Value  string `json:"value"`
	Format string `json:"format"`
	Min    string `json:"min"`
	Max    string `json:"max"`
	Step   string `json:"step"`
}

// LightProperty is a light property on a device. Note that these properties are read-only.
type LightProperty struct {
	Name        string                `json:"name"`
	Label       string                `json:"label"`
	Group       string                `json:"group"`
	State       PropertyState         `json:"state"`
	LastUpdated time.Time             `json:"lastUpdated"`
	Messages    []Message             `json:"messages"`
	Values      map[string]LightValue `json:"values"`
}

// LightValue is a light value on a LightProperty.
type LightValue struct {
	Name  string        `json:"name"`
	Label string        `json:"label"`
	Value PropertyState `json:"value"`
}

// BlobProperty is a blob property on a device.
type BlobProperty struct {
	Name        string               `json:"name"`
	Label       string               `json:"label"`
	Group       string               `json:"group"`
	State       PropertyState        `json:"state"`
	LastUpdated time.Time            `json:"lastUpdated"`
	Messages    []Message            `json:"messages"`
	Permissions PropertyPermission   `json:"permissions"`
	Timeout     int                  `json:"timeout"`
	Values      map[string]BlobValue `json:"values"`
}

// BlobValue is a blob value on a BlobProperty.
type BlobValue struct {
	Name  string `json:"name"`
	Label string `json:"label"`
	Value string `json:"value"`
	Size  int64  `json:"size"`
}

// Groups retreives a list of all the groups for a device for display purposes. Groups are returned in alphabetical order.
func (d Device) Groups() []string {
	temp := map[string]bool{}

	for _, v := range d.LightProperties {
		temp[v.Group] = true
	}

	for _, v := range d.TextProperties {
		temp[v.Group] = true
	}

	for _, v := range d.SwitchProperties {
		temp[v.Group] = true
	}

	for _, v := range d.BlobProperties {
		temp[v.Group] = true
	}

	for _, v := range d.NumberProperties {
		temp[v.Group] = true
	}

	groups := []string{}

	for k := range temp {
		groups = append(groups, k)
	}

	sort.Strings(groups)

	return groups
}
