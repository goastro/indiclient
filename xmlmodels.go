package indiclient

import (
	"encoding/xml"
)

type getProperties struct {
	XMLName xml.Name `xml:"getProperties"`
	Version string   `xml:"version,attr"`
	Device  string   `xml:"device,attr,omitempty"`
	Name    string   `xml:"name,attr,omitempty"`
}

// defTextVector
// Define a property that holds one or more text elements.
type defTextVector struct {
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
	Texts     []defText          `xml:"defText"`
}

// defText
// Define one member of a text vector.
type defText struct {
	XMLName xml.Name `xml:"defText"`
	Name    string   `xml:"name,attr"`
	Label   string   `xml:"label,attr"`
	Value   string   `xml:",chardata"`
}

// defNumberVector
// Define a property that holds one or more numeric values.
type defNumberVector struct {
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
	Numbers   []defNumber        `xml:"defNumber"`
}

// defNumber
// Define one member of a number vector.
type defNumber struct {
	XMLName xml.Name `xml:"defNumber"`
	Name    string   `xml:"name,attr"`
	Label   string   `xml:"label,attr"`
	Format  string   `xml:"format,attr"`
	Min     string   `xml:"min,attr"`
	Max     string   `xml:"max,attr"`
	Step    string   `xml:"step,attr"`
	Value   string   `xml:",chardata"`
}

// defSwitchVector
// Define a collection of switches. Rule is only a hint for use by a GUI to decide a suitable
// presentation style. Rules are actually implemented wholly within the Device.
type defSwitchVector struct {
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
	Switches  []defSwitch        `xml:"defSwitch"`
}

// defSwitch
// Define one member of a switch vector.
type defSwitch struct {
	XMLName xml.Name    `xml:"defSwitch"`
	Name    string      `xml:"name,attr"`
	Label   string      `xml:"label,attr"`
	Value   SwitchState `xml:",chardata"`
}

// defLightVector
// Define a collection of passive indicator lights.
type defLightVector struct {
	XMLName   xml.Name      `xml:"defLightVector"`
	Device    string        `xml:"device,attr"`
	Name      string        `xml:"name,attr"`
	Label     string        `xml:"label,attr"`
	Group     string        `xml:"group,attr"`
	State     PropertyState `xml:"state,attr"`
	Timestamp string        `xml:"timestamp,attr"`
	Message   string        `xml:"message"`
	Lights    []defLight    `xml:"defLight"`
}

// defLight
// Define one member of a light vector.
type defLight struct {
	XMLName xml.Name      `xml:"defLight"`
	Name    string        `xml:"name,attr"`
	Label   string        `xml:"label,attr"`
	Value   PropertyState `xml:",chardata"`
}

// defBlobVector
// Define a property that holds one or more Binary Large Objects, BLOBs.
type defBlobVector struct {
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
	Blobs     []defBlob          `xml:"defBLOB"`
}

// defBlob
// Define one member of a BLOB vector. Unlike other defXXX elements, this does not contain an
// initial value for the BLOB.
type defBlob struct {
	XMLName xml.Name `xml:"defBLOB"`
	Name    string   `xml:"name,attr"`
	Label   string   `xml:"label,attr"`
}

// enableBlob
// Command to control whether setBLOBs should be sent to this channel from a given Device. They can
// be turned off completely by setting Never (the default), allowed to be intermixed with other INDI
// commands by setting Also or made the only command by setting Only.
type enableBlob struct {
	XMLName xml.Name   `xml:"enableBLOB"`
	Device  string     `xml:"device,attr"`
	Name    string     `xml:"name,attr"`
	Value   BlobEnable `xml:",chardata"`
}

// newTextVector
// Commands to inform Device of new target values for a Property. After sending, the Client must set
// its local state for the Property to Busy, leaving it up to the Device to change it when it sees
// fit.
type newTextVector struct {
	XMLName   xml.Name  `xml:"newTextVector"`
	Device    string    `xml:"device,attr"`
	Name      string    `xml:"name,attr"`
	Timestamp string    `xml:"timestamp,attr,omitempty"`
	Texts     []oneText `xml:"oneText"`
}

// newNumberVector
// Commands to inform Device of new target values for a Property. After sending, the Client must set
// its local state for the Property to Busy, leaving it up to the Device to change it when it sees
// fit.
type newNumberVector struct {
	XMLName   xml.Name    `xml:"newNumberVector"`
	Device    string      `xml:"device,attr"`
	Name      string      `xml:"name,attr"`
	Timestamp string      `xml:"timestamp,attr,omitempty"`
	Numbers   []oneNumber `xml:"oneNumber"`
}

// newSwitchVector
// Commands to inform Device of new target values for a Property. After sending, the Client must set
// its local state for the Property to Busy, leaving it up to the Device to change it when it sees
// fit.
type newSwitchVector struct {
	XMLName   xml.Name    `xml:"newSwitchVector"`
	Device    string      `xml:"device,attr"`
	Name      string      `xml:"name,attr"`
	Timestamp string      `xml:"timestamp,attr,omitempty"`
	Switches  []oneSwitch `xml:"oneSwitch"`
}

// newBlobVector
// Commands to inform Device of new target values for a Property. After sending, the Client must set
// its local state for the Property to Busy, leaving it up to the Device to change it when it sees
// fit.
type newBlobVector struct {
	XMLName   xml.Name  `xml:"newBLOBVector"`
	Device    string    `xml:"device,attr"`
	Name      string    `xml:"name,attr"`
	Timestamp string    `xml:"timestamp,attr,omitempty"`
	Blobs     []oneBlob `xml:"oneBLOB"`
}

// oneText
// One member of a Text vector.
type oneText struct {
	XMLName xml.Name `xml:"oneText"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:",chardata"`
}

// oneNumber
// One member of a Number vector.
type oneNumber struct {
	XMLName xml.Name `xml:"oneNumber"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:",chardata"`
}

// oneSwitch
// One member of a switch vector.
type oneSwitch struct {
	XMLName xml.Name    `xml:"oneSwitch"`
	Name    string      `xml:"name,attr"`
	Value   SwitchState `xml:",chardata"`
}

// oneBlob
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
type oneBlob struct {
	XMLName xml.Name `xml:"oneBLOB"`
	Name    string   `xml:"name,attr"`
	Size    int      `xml:"size,attr"`
	Format  string   `xml:"format,attr"`
	Value   string   `xml:",chardata"`
}

// oneLight
// Send a message to specify state of one member of a Light vector
type oneLight struct {
	XMLName xml.Name      `xml:"oneLight"`
	Name    string        `xml:"name,attr"`
	Value   PropertyState `xml:",chardata"`
}

// setTextVector
// Send a new set of values for a Text vector, with optional new timeout, state and message.
type setTextVector struct {
	XMLName   xml.Name      `xml:"setTextVector"`
	Device    string        `xml:"device,attr"`
	Name      string        `xml:"name,attr"`
	State     PropertyState `xml:"state,attr"`
	Timeout   int           `xml:"timeout,attr"`
	Timestamp string        `xml:"timestamp,attr"`
	Message   string        `xml:"message"`
	Texts     []oneText     `xml:"oneText"`
}

// setNumberVector
// Send a new set of values for a Number vector, with optional new timeout, state and message.
type setNumberVector struct {
	XMLName   xml.Name      `xml:"setNumberVector"`
	Device    string        `xml:"device,attr"`
	Name      string        `xml:"name,attr"`
	State     PropertyState `xml:"state,attr"`
	Timeout   int           `xml:"timeout,attr"`
	Timestamp string        `xml:"timestamp,attr"`
	Message   string        `xml:"message"`
	Numbers   []oneNumber   `xml:"oneNumber"`
}

// setSwitchVector
// Send a new set of values for a Switch vector, with optional new timeout, state and message.
type setSwitchVector struct {
	XMLName   xml.Name      `xml:"setSwitchVector"`
	Device    string        `xml:"device,attr"`
	Name      string        `xml:"name,attr"`
	State     PropertyState `xml:"state,attr"`
	Timeout   int           `xml:"timeout,attr"`
	Timestamp string        `xml:"timestamp,attr"`
	Message   string        `xml:"message"`
	Switches  []oneSwitch   `xml:"oneSwitch"`
}

// setLightVector
// Send a new set of values for a Light vector, with optional new state and message.
type setLightVector struct {
	XMLName   xml.Name      `xml:"setLightVector"`
	Device    string        `xml:"device,attr"`
	Name      string        `xml:"name,attr"`
	State     PropertyState `xml:"state,attr"`
	Timestamp string        `xml:"timestamp,attr"`
	Message   string        `xml:"message"`
	Lights    []oneLight    `xml:"oneLight"`
}

// setBlobVector
// Send a new set of values for a BLOB vector, with optional new timeout, state and message.
type setBlobVector struct {
	XMLName   xml.Name      `xml:"setBLOBVector"`
	Device    string        `xml:"device,attr"`
	Name      string        `xml:"name,attr"`
	State     PropertyState `xml:"state,attr"`
	Timeout   int           `xml:"timeout,attr"`
	Timestamp string        `xml:"timestamp,attr"`
	Message   string        `xml:"message"`
	Blobs     []oneBlob     `xml:"oneBLOB"`
}

// message
// Send a message associated with a device or entire system.
type message struct {
	XMLName   xml.Name `xml:"message"`
	Device    string   `xml:"device,attr"`
	Timestamp string   `xml:"timestamp,attr"`
	Message   string   `xml:"message,attr"`
}

// delProperty
// Delete the given property, or entire device if no property is specified.
type delProperty struct {
	XMLName   xml.Name `xml:"delProperty"`
	Device    string   `xml:"device,attr"`
	Name      string   `xml:"name,attr"`
	Timestamp string   `xml:"timestamp,attr"`
	Message   string   `xml:"message,attr"`
}
