package indiclient

import (
	"encoding/xml"
)

type GetProperties struct {
	XMLName xml.Name `xml:"getProperties"`
	Version string   `xml:"version,attr"`
	Device  string   `xml:"device,attr,omitempty"`
	Name    string   `xml:"name,attr,omitempty"`
}

// DefTextVector
// Define a property that holds one or more text elements.
type DefTextVector struct {
	XMLName   xml.Name           `xml:"defTextVector"`
	Device    string             `xml:"device,attr"`
	Name      string             `xml:"name,attr"`
	Label     string             `xml:"label,attr"`
	Group     string             `xml:"group,attr"`
	State     PropertyState      `xml:"state,attr"`
	Perm      PropertyPermission `xml:"perm,attr"`
	Timeout   int                `xml:"timeout,attr"`
	Timestamp string             `xml:"timestamp,attr"`
	Message   string             `xml:"message"`
	Texts     []DefText          `xml:"defText"`
}

// DefText
// Define one member of a text vector.
type DefText struct {
	XMLName xml.Name `xml:"defText"`
	Name    string   `xml:"name,attr"`
	Label   string   `xml:"label,attr"`
	Value   string   `xml:",chardata"`
}

// DefNumberVector
// Define a property that holds one or more numeric values.
type DefNumberVector struct {
	XMLName   xml.Name           `xml:"defNumberVector"`
	Device    string             `xml:"device,attr"`
	Name      string             `xml:"name,attr"`
	Label     string             `xml:"label,attr"`
	Group     string             `xml:"group,attr"`
	State     PropertyState      `xml:"state,attr"`
	Perm      PropertyPermission `xml:"perm,attr"`
	Timeout   int                `xml:"timeout,attr"`
	Timestamp string             `xml:"timestamp,attr"`
	Message   string             `xml:"message"`
	Numbers   []DefNumber        `xml:"defNumber"`
}

// DefNumber
// Define one member of a number vector.
type DefNumber struct {
	XMLName xml.Name `xml:"defNumber"`
	Name    string   `xml:"name,attr"`
	Label   string   `xml:"label,attr"`
	Format  string   `xml:"format,attr"`
	Min     string   `xml:"min,attr"`
	Max     string   `xml:"max,attr"`
	Step    string   `xml:"step,attr"`
	Value   string   `xml:",chardata"`
}

// DefSwitchVector
// Define a collection of switches. Rule is only a hint for use by a GUI to decide a suitable
// presentation style. Rules are actually implemented wholly within the Device.
type DefSwitchVector struct {
	XMLName   xml.Name           `xml:"defSwitchVector"`
	Device    string             `xml:"device,attr"`
	Name      string             `xml:"name,attr"`
	Label     string             `xml:"label,attr"`
	Group     string             `xml:"group,attr"`
	State     PropertyState      `xml:"state,attr"`
	Perm      PropertyPermission `xml:"perm,attr"`
	Rule      SwitchRule         `xml:"rule,attr"`
	Timeout   int                `xml:"timeout,attr"`
	Timestamp string             `xml:"timestamp,attr"`
	Message   string             `xml:"message"`
	Switches  []DefSwitch        `xml:"defSwitch"`
}

// DefSwitch
// Define one member of a switch vector.
type DefSwitch struct {
	XMLName xml.Name    `xml:"defSwitch"`
	Name    string      `xml:"name,attr"`
	Label   string      `xml:"label,attr"`
	Value   SwitchState `xml:",chardata"`
}

// DefLightVector
// Define a collection of passive indicator lights.
type DefLightVector struct {
	XMLName   xml.Name      `xml:"defLightVector"`
	Device    string        `xml:"device,attr"`
	Name      string        `xml:"name,attr"`
	Label     string        `xml:"label,attr"`
	Group     string        `xml:"group,attr"`
	State     PropertyState `xml:"state,attr"`
	Timestamp string        `xml:"timestamp,attr"`
	Message   string        `xml:"message"`
	Lights    []DefLight    `xml:"defLight"`
}

// DefLight
// Define one member of a light vector.
type DefLight struct {
	XMLName xml.Name      `xml:"defLight"`
	Name    string        `xml:"name,attr"`
	Label   string        `xml:"label,attr"`
	Value   PropertyState `xml:",chardata"`
}

// DefBlobVector
// Define a property that holds one or more Binary Large Objects, BLOBs.
type DefBlobVector struct {
	XMLName   xml.Name           `xml:"defBLOBVector"`
	Device    string             `xml:"device,attr"`
	Name      string             `xml:"name,attr"`
	Label     string             `xml:"label,attr"`
	Group     string             `xml:"group,attr"`
	State     PropertyState      `xml:"state,attr"`
	Perm      PropertyPermission `xml:"perm,attr"`
	Timeout   int                `xml:"timeout,attr"`
	Timestamp string             `xml:"timestamp,attr"`
	Message   string             `xml:"message"`
	Blobs     []DefBlob          `xml:"defBLOB"`
}

// DefBlob
// Define one member of a BLOB vector. Unlike other defXXX elements, this does not contain an
// initial value for the BLOB.
type DefBlob struct {
	XMLName xml.Name `xml:"defBLOB"`
	Name    string   `xml:"name,attr"`
	Label   string   `xml:"label,attr"`
}

// EnableBlob
// Command to control whether setBLOBs should be sent to this channel from a given Device. They can
// be turned off completely by setting Never (the default), allowed to be intermixed with other INDI
// commands by setting Also or made the only command by setting Only.
type EnableBlob struct {
	XMLName xml.Name   `xml:"enableBLOB"`
	Device  string     `xml:"device,attr"`
	Name    string     `xml:"name,attr"`
	Value   BlobEnable `xml:",chardata"`
}

// NewTextVector
// Commands to inform Device of new target values for a Property. After sending, the Client must set
// its local state for the Property to Busy, leaving it up to the Device to change it when it sees
// fit.
type NewTextVector struct {
	XMLName   xml.Name  `xml:"newTextVector"`
	Device    string    `xml:"device,attr"`
	Name      string    `xml:"name,attr"`
	Timestamp string    `xml:"timestamp,attr,omitempty"`
	Texts     []OneText `xml:"oneText"`
}

// NewNumberVector
// Commands to inform Device of new target values for a Property. After sending, the Client must set
// its local state for the Property to Busy, leaving it up to the Device to change it when it sees
// fit.
type NewNumberVector struct {
	XMLName   xml.Name    `xml:"newNumberVector"`
	Device    string      `xml:"device,attr"`
	Name      string      `xml:"name,attr"`
	Timestamp string      `xml:"timestamp,attr,omitempty"`
	Numbers   []OneNumber `xml:"oneNumber"`
}

// NewSwitchVector
// Commands to inform Device of new target values for a Property. After sending, the Client must set
// its local state for the Property to Busy, leaving it up to the Device to change it when it sees
// fit.
type NewSwitchVector struct {
	XMLName   xml.Name    `xml:"newSwitchVector"`
	Device    string      `xml:"device,attr"`
	Name      string      `xml:"name,attr"`
	Timestamp string      `xml:"timestamp,attr,omitempty"`
	Switches  []OneSwitch `xml:"oneSwitch"`
}

// NewBlobVector
// Commands to inform Device of new target values for a Property. After sending, the Client must set
// its local state for the Property to Busy, leaving it up to the Device to change it when it sees
// fit.
type NewBlobVector struct {
	XMLName   xml.Name  `xml:"newBLOBVector"`
	Device    string    `xml:"device,attr"`
	Name      string    `xml:"name,attr"`
	Timestamp string    `xml:"timestamp,attr,omitempty"`
	Blobs     []OneBlob `xml:"oneBLOB"`
}

// OneText
// One member of a Text vector.
type OneText struct {
	XMLName xml.Name `xml:"oneText"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:",chardata"`
}

// OneNumber
// One member of a Number vector.
type OneNumber struct {
	XMLName xml.Name `xml:"oneNumber"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:",chardata"`
}

// OneSwitch
// One member of a switch vector.
type OneSwitch struct {
	XMLName xml.Name    `xml:"oneSwitch"`
	Name    string      `xml:"name,attr"`
	Value   SwitchState `xml:",chardata"`
}

// OneBlob
// One member of a BLOB vector. The contents of this element must always be encoded using base64.
// The format attribute consists of one or more file name suffixes, each preceded with a period,
// which indicate how the decoded data is to be interpreted. For example .fits indicates the decoded
// BLOB is a FITS file, and .fits.z indicates the decoded BLOB is a FITS file compressed with
// zlib. The INDI protocol places no restrictions on the contents or formats of BLOBs but at
// minimum astronomical INDI clients are encouraged to support the FITS image file format and the
// zlib compression mechanism. The size attribute indicates the number of bytes in the final BLOB
// after decoding and after any decompression. For example, if the format is .fits.z the size
// attribute is the number of bytes in the FITS file. A Client unfamiliar with the specified format
// may use the attribute as a simple string, perhaps in combination with the timestamp attribute, to
// create a file name in which to store the data without processing other than decoding the base64.
type OneBlob struct {
	XMLName xml.Name `xml:"oneBLOB"`
	Name    string   `xml:"name,attr"`
	Size    int      `xml:"size,attr"`
	Format  string   `xml:"format,attr"`
	Value   string   `xml:",chardata"`
}

// OneLight
// Send a message to specify state of one member of a Light vector
type OneLight struct {
	XMLName xml.Name      `xml:"oneLight"`
	Name    string        `xml:"name,attr"`
	Value   PropertyState `xml:",chardata"`
}

// SetTextVector
// Send a new set of values for a Text vector, with optional new timeout, state and message.
type SetTextVector struct {
	XMLName   xml.Name      `xml:"setTextVector"`
	Device    string        `xml:"device,attr"`
	Name      string        `xml:"name,attr"`
	State     PropertyState `xml:"state,attr"`
	Timeout   int           `xml:"timeout,attr"`
	Timestamp string        `xml:"timestamp,attr"`
	Message   string        `xml:"message"`
	Texts     []OneText     `xml:"oneText"`
}

// SetNumberVector
// Send a new set of values for a Number vector, with optional new timeout, state and message.
type SetNumberVector struct {
	XMLName   xml.Name      `xml:"setNumberVector"`
	Device    string        `xml:"device,attr"`
	Name      string        `xml:"name,attr"`
	State     PropertyState `xml:"state,attr"`
	Timeout   int           `xml:"timeout,attr"`
	Timestamp string        `xml:"timestamp,attr"`
	Message   string        `xml:"message"`
	Numbers   []OneNumber   `xml:"oneNumber"`
}

// SetSwitchVector
// Send a new set of values for a Switch vector, with optional new timeout, state and message.
type SetSwitchVector struct {
	XMLName   xml.Name      `xml:"setSwitchVector"`
	Device    string        `xml:"device,attr"`
	Name      string        `xml:"name,attr"`
	State     PropertyState `xml:"state,attr"`
	Timeout   int           `xml:"timeout,attr"`
	Timestamp string        `xml:"timestamp,attr"`
	Message   string        `xml:"message"`
	Switches  []OneSwitch   `xml:"oneSwitch"`
}

// SetLightVector
// Send a new set of values for a Light vector, with optional new state and message.
type SetLightVector struct {
	XMLName   xml.Name      `xml:"setLightVector"`
	Device    string        `xml:"device,attr"`
	Name      string        `xml:"name,attr"`
	State     PropertyState `xml:"state,attr"`
	Timestamp string        `xml:"timestamp,attr"`
	Message   string        `xml:"message"`
	Lights    []OneLight    `xml:"oneLight"`
}

// SetBlobVector
// Send a new set of values for a BLOB vector, with optional new timeout, state and message.
type SetBlobVector struct {
	XMLName   xml.Name      `xml:"setBLOBVector"`
	Device    string        `xml:"device,attr"`
	Name      string        `xml:"name,attr"`
	State     PropertyState `xml:"state,attr"`
	Timeout   int           `xml:"timeout,attr"`
	Timestamp string        `xml:"timestamp,attr"`
	Message   string        `xml:"message"`
	Blobs     []OneBlob     `xml:"oneBLOB"`
}

// Message
// Send a message associated with a device or entire system.
type Message struct {
	XMLName   xml.Name `xml:"message"`
	Device    string   `xml:"device,attr"`
	Timestamp string   `xml:"timestamp,attr"`
	Message   string   `xml:"message,attr"`
}

// DelProperty
// Delete the given property, or entire device if no property is specified.
type DelProperty struct {
	XMLName   xml.Name `xml:"delProperty"`
	Device    string   `xml:"device,attr"`
	Name      string   `xml:"name,attr"`
	Timestamp string   `xml:"timestamp,attr"`
	Message   string   `xml:"message,attr"`
}
