// Package indiclient is a pure Go implementation of an indi client. It supports indiserver version 1.7.
//
// See http://indilib.org/develop/developer-manual/106-client-development.html
//
// See http://www.clearskyinstitute.com/INDI/INDI.pdf
//
// One of the awesome, but sometimes infuriating features of the INDI protocol is that if a device receives
// a command it doesn't understand, it is under no obligation to respond, and usually won't. This can make
// debugging difficult, because you aren't always sure if you are just sending the command incorrectly or
// if there is something else wrong. This library tries to alleviate that by checking parameters to all
// calls and will return an error if something doesn't look right.
package indiclient

// TODO: Handle device timeouts

import (
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rickbassham/logging"
	"github.com/spf13/afero"
)

var (
	// ErrDeviceNotFound is returned when a call cannot find a device.
	ErrDeviceNotFound = errors.New("device not found")

	// ErrPropertyNotFound is returned when a call cannot find a property.
	ErrPropertyNotFound = errors.New("property not found")

	// ErrPropertyValueNotFound is returned when a call cannot find a property value.
	ErrPropertyValueNotFound = errors.New("property value not found")

	// ErrPropertyReadOnly is returned when an attempt to change a read-only property was made.
	ErrPropertyReadOnly = errors.New("property read only")

	// ErrPropertyWithoutDevice is returned when an attempt to GetProperties specifies a property but no device.
	ErrPropertyWithoutDevice = errors.New("property specified without device")

	// ErrInvalidBlobEnable is returned when a value other than Only, Also, Never is specified for BlobEnable.
	ErrInvalidBlobEnable = errors.New("invalid BlobEnable value")
)

// PropertyState represents the current state of a property. "Idle", "Ok", "Busy", or "Alert".
type PropertyState string

const (
	// PropertyStateIdle represents a property that is "Idle". This is recommended to be displayed as Gray.
	PropertyStateIdle = PropertyState("Idle")
	// PropertyStateOk represents a property that is "Ok". This is recommended to be displayed as Green.
	PropertyStateOk = PropertyState("Ok")
	// PropertyStateBusy represents a property that is "Busy". This is recommended to be displayed as Yellow.
	PropertyStateBusy = PropertyState("Busy")
	// PropertyStateAlert represents a property that is "Alert". This is recommended to be displayed as Red.
	PropertyStateAlert = PropertyState("Alert")
)

// SwitchState reprensents the current state of a switch value. "On" or "Off".
type SwitchState string

const (
	// SwitchStateOff represents a switch that is "Off".
	SwitchStateOff = SwitchState("Off")
	// SwitchStateOn represents a switch that is "On".
	SwitchStateOn = SwitchState("On")
)

// SwitchRule represents how a switch state can exist relative to the other switches in the vector. "OneOfMany", "AtMostOne", or "AnyOfMany".
type SwitchRule string

const (
	// SwitchRuleOneOfMany represents a switch that must have one switch in a vector active at a time.
	SwitchRuleOneOfMany = SwitchRule("OneOfMany")
	// SwitchRuleAtMostOne represents a switch that must have no more than one switch in a vector active at a time.
	SwitchRuleAtMostOne = SwitchRule("AtMostOne")
	// SwitchRuleAnyOfMany represents a switch that may have any number of switches in a vector active at a time.
	SwitchRuleAnyOfMany = SwitchRule("AnyOfMany")
)

// PropertyPermission represents a permission hint for the client. "ro", "wo", or "rw".
type PropertyPermission string

const (
	// PropertyPermissionReadOnly represents a property that is Read-Only.
	PropertyPermissionReadOnly = PropertyPermission("ro")
	// PropertyPermissionWriteOnly represents a property that is Write-Only.
	PropertyPermissionWriteOnly = PropertyPermission("wo")
	// PropertyPermissionReadWrite represents a property that is Read-Write.
	PropertyPermissionReadWrite = PropertyPermission("rw")
)

// BlobEnable represents whether BLOB's should be sent to this client.
type BlobEnable string

const (
	// BlobEnableNever (default) represents that the current client should not be sent any BLOB's for a device.
	BlobEnableNever = BlobEnable("Never")
	// BlobEnableAlso represents that the current client should be sent any BLOB's for a device in addition to the normal INDI commands.
	BlobEnableAlso = BlobEnable("Also")
	// BlobEnableOnly represents that the current client should only be sent any BLOB's for a device.
	BlobEnableOnly = BlobEnable("Only")
)

// Dialer allows the client to connect to an INDI server.
type Dialer interface {
	Dial(network, address string) (io.ReadWriteCloser, error)
}

// NetworkDialer is an implementation of Dialer that uses the built-in net package.
type NetworkDialer struct{}

// Dial connects to the address on the named network.
func (NetworkDialer) Dial(network, address string) (io.ReadWriteCloser, error) {
	return net.Dial(network, address)
}

// INDIClient is the struct used to keep a connection alive to an indiserver.
type INDIClient struct {
	log        logging.Logger
	dialer     Dialer
	fs         afero.Fs
	bufferSize int

	conn io.ReadWriteCloser

	write chan interface{}
	read  chan interface{}

	devices     sync.Map
	blobStreams sync.Map
}

// NewINDIClient creates a client to connect to an INDI server.
func NewINDIClient(log logging.Logger, dialer Dialer, fs afero.Fs, bufferSize int) *INDIClient {
	return &INDIClient{
		log:         log,
		dialer:      dialer,
		devices:     sync.Map{},
		blobStreams: sync.Map{},
		fs:          fs,
		bufferSize:  bufferSize,
	}
}

// Connect dials to create a connection to address. address should be in the format that the provided Dialer expects.
func (c *INDIClient) Connect(network, address string) error {
	conn, err := c.dialer.Dial(network, address)
	if err != nil {
		return err
	}

	// Clear out all devices
	c.delProperty(&delProperty{})

	c.conn = conn

	c.read = make(chan interface{}, c.bufferSize)
	c.write = make(chan interface{}, c.bufferSize)

	c.startRead()
	c.startWrite()

	return nil
}

// Disconnect clears out all devices from memory, closes the connection, and closes the read and write channels.
func (c *INDIClient) Disconnect() error {
	// Clear out all devices
	c.delProperty(&delProperty{})

	if c.conn == nil {
		return nil
	}

	err := c.conn.Close()
	c.conn = nil

	if c.read != nil {
		close(c.read)
	}

	if c.write != nil {
		close(c.write)
	}

	return err
}

// IsConnected returns true if the client is currently connected to an INDI server. Otherwise, returns false.
func (c *INDIClient) IsConnected() bool {
	if c.conn != nil {
		return true
	}

	return false
}

// Devices returns the current list of INDI devices with their current state.
func (c *INDIClient) Devices() []Device {
	devices := []Device{}

	c.devices.Range(func(key, value interface{}) bool {
		devices = append(devices, value.(Device))

		return true
	})

	return devices
}

// GetBlob finds a BLOB with the given deviceName, propName, blobName. Be sure to close rdr when you are done with it.
func (c *INDIClient) GetBlob(deviceName, propName, blobName string) (rdr io.ReadCloser, fileName string, length int64, err error) {
	device, err := c.findDevice(deviceName)
	if err != nil {
		return
	}

	prop, ok := device.BlobProperties[propName]
	if !ok {
		err = ErrPropertyNotFound
		return
	}

	val, ok := prop.Values[blobName]
	if !ok {
		err = ErrPropertyValueNotFound
		return
	}

	rdr, err = c.fs.Open(val.Value)
	if err != nil {
		return
	}

	fileName = filepath.Base(val.Value)

	length = val.Size
	return
}

// GetBlobStream finds a BLOB with the given deviceName, propName, blobName. This will return an io.Pipe that can stream the BLOBs that are received from the indiserver.
// The client will keep track of all open streams and write to them as blobs are received from indiserver. Remember to call CloseBlobStream when you are done. If you don't,
// all blobs received for that device, property, blob will fail to write once the reader is closed.
func (c *INDIClient) GetBlobStream(deviceName, propName, blobName string) (rdr io.ReadCloser, id string, err error) {
	device, err := c.findDevice(deviceName)
	if err != nil {
		return
	}

	prop, ok := device.BlobProperties[propName]
	if !ok {
		err = ErrPropertyNotFound
		return
	}

	_, ok = prop.Values[blobName]
	if !ok {
		err = ErrPropertyValueNotFound
		return
	}

	guid := uuid.New()
	id = guid.String()

	key := fmt.Sprintf("%s_%s_%s", deviceName, propName, blobName)

	r, w := io.Pipe()

	rdr = r

	writers := map[string]io.Writer{}

	if ws, ok := c.blobStreams.Load(key); ok {
		writers = ws.(map[string]io.Writer)
	}

	writers[id] = w

	c.blobStreams.Store(key, writers)

	return
}

// CloseBlobStream closes the blob stream created by GetBlobStream.
func (c *INDIClient) CloseBlobStream(deviceName, propName, blobName string, id string) (err error) {
	device, err := c.findDevice(deviceName)
	if err != nil {
		return
	}

	prop, ok := device.BlobProperties[propName]
	if !ok {
		err = ErrPropertyNotFound
		return
	}

	_, ok = prop.Values[blobName]
	if !ok {
		err = ErrPropertyValueNotFound
		return
	}
	key := fmt.Sprintf("%s_%s_%s", deviceName, propName, blobName)

	if ws, ok := c.blobStreams.Load(key); ok {
		writers := ws.(map[string]io.Writer)

		if w, ok := writers[id]; ok {
			w.(io.WriteCloser).Close()

			delete(writers, id)

			c.blobStreams.Store(key, writers)
		}
	}

	return
}

// GetProperties sends a command to the INDI server to retreive the property definitions for the given deviceName and propName.
// deviceName and propName are optional.
func (c *INDIClient) GetProperties(deviceName, propName string) error {
	if len(propName) > 0 && len(deviceName) == 0 {
		return ErrPropertyWithoutDevice
	}

	cmd := getProperties{
		Version: "1.7",
		Device:  deviceName,
		Name:    propName,
	}

	c.write <- cmd

	return nil
}

// EnableBlob sends a command to the INDI server to enable/disable BLOBs for the current connection.
// It is recommended to enable blobs on their own client, and keep the main connection clear of large transfers.
// By default, BLOBs are NOT enabled.
func (c *INDIClient) EnableBlob(deviceName, propName string, val BlobEnable) error {
	if val != BlobEnableAlso && val != BlobEnableNever && val != BlobEnableOnly {
		return ErrInvalidBlobEnable
	}

	_, err := c.findDevice(deviceName)
	if err != nil {
		return err
	}

	cmd := enableBlob{
		Device: deviceName,
		Name:   propName,
		Value:  val,
	}

	c.write <- cmd

	return nil
}

// SetTextValue sends a command to the INDI server to change the value of a textVector.
func (c *INDIClient) SetTextValue(deviceName, propName, textName, textValue string) error {
	device, err := c.findDevice(deviceName)
	if err != nil {
		return err
	}

	prop, ok := device.TextProperties[propName]
	if !ok {
		return ErrPropertyNotFound
	}

	if prop.Permissions == PropertyPermissionReadOnly {
		return ErrPropertyReadOnly
	}

	_, ok = prop.Values[textName]
	if !ok {
		return ErrPropertyValueNotFound
	}

	prop.State = PropertyStateBusy

	device.TextProperties[propName] = prop

	c.devices.Store(deviceName, device)

	cmd := newTextVector{
		Device: deviceName,
		Name:   propName,
		Texts: []oneText{
			{
				Name:  textName,
				Value: textValue,
			},
		},
	}

	c.write <- cmd

	return nil
}

// SetNumberValue sends a command to the INDI server to change the value of a numberVector.
func (c *INDIClient) SetNumberValue(deviceName, propName, NumberName, NumberValue string) error {
	device, err := c.findDevice(deviceName)
	if err != nil {
		return err
	}

	prop, ok := device.NumberProperties[propName]
	if !ok {
		return ErrPropertyNotFound
	}

	if prop.Permissions == PropertyPermissionReadOnly {
		return ErrPropertyReadOnly
	}

	_, ok = prop.Values[NumberName]
	if !ok {
		return ErrPropertyValueNotFound
	}

	prop.State = PropertyStateBusy

	device.NumberProperties[propName] = prop

	c.devices.Store(deviceName, device)

	cmd := newNumberVector{
		Device: deviceName,
		Name:   propName,
		Numbers: []oneNumber{
			{
				Name:  NumberName,
				Value: NumberValue,
			},
		},
	}

	c.write <- cmd

	return nil
}

// SetSwitchValue sends a command to the INDI server to change the value of a switchVector.
// Note that you will ususally set the desired property on SwitchStateOn, and let the device
// decide how to switch the other values off.
func (c *INDIClient) SetSwitchValue(deviceName, propName, switchName string, switchValue SwitchState) error {
	device, err := c.findDevice(deviceName)
	if err != nil {
		return err
	}

	prop, ok := device.SwitchProperties[propName]
	if !ok {
		return ErrPropertyNotFound
	}

	if prop.Permissions == PropertyPermissionReadOnly {
		return ErrPropertyReadOnly
	}

	_, ok = prop.Values[switchName]
	if !ok {
		return ErrPropertyValueNotFound
	}

	prop.State = PropertyStateBusy

	device.SwitchProperties[propName] = prop

	c.devices.Store(deviceName, device)

	cmd := newSwitchVector{
		Device: deviceName,
		Name:   propName,
		Switches: []oneSwitch{
			{
				Name:  switchName,
				Value: switchValue,
			},
		},
	}

	c.write <- cmd

	return nil
}

// SetBlobValue sends a command to the INDI server to change the value of a blobVector.
func (c *INDIClient) SetBlobValue(deviceName, propName, blobName, blobValue, blobFormat string, blobSize int) error {
	device, err := c.findDevice(deviceName)
	if err != nil {
		return err
	}

	prop, ok := device.BlobProperties[propName]
	if !ok {
		return ErrPropertyNotFound
	}

	if prop.Permissions == PropertyPermissionReadOnly {
		return ErrPropertyReadOnly
	}

	_, ok = prop.Values[blobName]
	if !ok {
		return ErrPropertyValueNotFound
	}

	prop.State = PropertyStateBusy

	device.BlobProperties[propName] = prop

	c.devices.Store(deviceName, device)

	cmd := newBlobVector{
		Device: deviceName,
		Name:   propName,
		Blobs: []oneBlob{
			{
				Name:   blobName,
				Value:  blobValue,
				Size:   blobSize,
				Format: blobFormat,
			},
		},
	}

	c.write <- cmd

	return nil
}

func (c *INDIClient) findDevice(name string) (Device, error) {
	if d, ok := c.devices.Load(name); ok {
		return d.(Device), nil
	}

	return Device{}, ErrDeviceNotFound
}

func (c *INDIClient) findOrCreateDevice(name string) Device {
	device, err := c.findDevice(name)
	if err == ErrDeviceNotFound {
		device = Device{
			Name:             name,
			TextProperties:   map[string]TextProperty{},
			SwitchProperties: map[string]SwitchProperty{},
			NumberProperties: map[string]NumberProperty{},
			LightProperties:  map[string]LightProperty{},
			BlobProperties:   map[string]BlobProperty{},
		}
	}

	return device
}

type indiMessageHandler interface {
	defTextVector(item *defTextVector)
	defSwitchVector(item *defSwitchVector)
	defNumberVector(item *defNumberVector)
	defLightVector(item *defLightVector)
	defBlobVector(item *defBlobVector)
	setSwitchVector(item *setSwitchVector)
	setTextVector(item *setTextVector)
	setNumberVector(item *setNumberVector)
	setLightVector(item *setLightVector)
	setBlobVector(item *setBlobVector)
	message(item *message)
	delProperty(item *delProperty)
}

func (c *INDIClient) defTextVector(item *defTextVector) {
	device := c.findOrCreateDevice(item.Device)

	prop := TextProperty{
		Name:        item.Name,
		Label:       item.Label,
		Group:       item.Group,
		Permissions: item.Perm,
		State:       item.State,
		Values:      map[string]TextValue{},
		LastUpdated: time.Now(),
		Messages:    []Message{},
	}

	for _, val := range item.Texts {
		prop.Values[val.Name] = TextValue{
			Label: val.Label,
			Name:  val.Name,
			Value: strings.TrimSpace(val.Value),
		}
	}

	if len(item.Message) > 0 {
		prop.Messages = append(prop.Messages, Message{
			Message:   item.Message,
			Timestamp: time.Now(),
		})
	}

	device.TextProperties[item.Name] = prop

	c.devices.Store(item.Device, device)
}

func (c *INDIClient) defSwitchVector(item *defSwitchVector) {
	device := c.findOrCreateDevice(item.Device)

	prop := SwitchProperty{
		Name:        item.Name,
		Label:       item.Label,
		Group:       item.Group,
		Permissions: item.Perm,
		Rule:        item.Rule,
		State:       item.State,
		Values:      map[string]SwitchValue{},
		LastUpdated: time.Now(),
		Messages:    []Message{},
	}

	for _, val := range item.Switches {
		prop.Values[val.Name] = SwitchValue{
			Label: val.Label,
			Name:  val.Name,
			Value: SwitchState(strings.TrimSpace(string(val.Value))),
		}
	}

	if len(item.Message) > 0 {
		prop.Messages = append(prop.Messages, Message{
			Message:   item.Message,
			Timestamp: time.Now(),
		})
	}

	device.SwitchProperties[item.Name] = prop

	c.devices.Store(item.Device, device)
}

func (c *INDIClient) defNumberVector(item *defNumberVector) {
	device := c.findOrCreateDevice(item.Device)

	prop := NumberProperty{
		Name:        item.Name,
		Label:       item.Label,
		Group:       item.Group,
		Permissions: item.Perm,
		State:       item.State,
		Values:      map[string]NumberValue{},
		LastUpdated: time.Now(),
		Messages:    []Message{},
	}

	for _, val := range item.Numbers {
		prop.Values[val.Name] = NumberValue{
			Label:  val.Label,
			Name:   val.Name,
			Value:  strings.TrimSpace(val.Value),
			Format: val.Format,
			Min:    val.Min,
			Max:    val.Max,
			Step:   val.Step,
		}
	}

	if len(item.Message) > 0 {
		prop.Messages = append(prop.Messages, Message{
			Message:   item.Message,
			Timestamp: time.Now(),
		})
	}

	device.NumberProperties[item.Name] = prop

	c.devices.Store(item.Device, device)
}

func (c *INDIClient) defLightVector(item *defLightVector) {
	device := c.findOrCreateDevice(item.Device)

	prop := LightProperty{
		Name:        item.Name,
		Label:       item.Label,
		Group:       item.Group,
		State:       item.State,
		Values:      map[string]LightValue{},
		LastUpdated: time.Now(),
		Messages:    []Message{},
	}

	for _, val := range item.Lights {
		prop.Values[val.Name] = LightValue{
			Label: val.Label,
			Name:  val.Name,
			Value: PropertyState(strings.TrimSpace(string(val.Value))),
		}
	}

	if len(item.Message) > 0 {
		prop.Messages = append(prop.Messages, Message{
			Message:   item.Message,
			Timestamp: time.Now(),
		})
	}

	device.LightProperties[item.Name] = prop

	c.devices.Store(item.Device, device)
}

func (c *INDIClient) defBlobVector(item *defBlobVector) {
	device := c.findOrCreateDevice(item.Device)

	prop := BlobProperty{
		Name:        item.Name,
		Label:       item.Label,
		Group:       item.Group,
		State:       item.State,
		Values:      map[string]BlobValue{},
		LastUpdated: time.Now(),
		Messages:    []Message{},
	}

	for _, val := range item.Blobs {
		prop.Values[val.Name] = BlobValue{
			Label: val.Label,
			Name:  val.Name,
		}
	}

	if len(item.Message) > 0 {
		prop.Messages = append(prop.Messages, Message{
			Message:   item.Message,
			Timestamp: time.Now(),
		})
	}

	device.BlobProperties[item.Name] = prop

	c.devices.Store(item.Device, device)
}

func (c *INDIClient) setSwitchVector(item *setSwitchVector) {
	device, err := c.findDevice(item.Device)
	if err != nil {
		c.log.WithField("device", item.Device).WithError(err).Warn("could not find device")
		return
	}

	var prop SwitchProperty
	if p, ok := device.SwitchProperties[item.Name]; ok {
		prop = p
	} else {
		c.log.WithField("device", item.Device).WithField("property", item.Name).Warn("could not find property")
		return
	}

	prop.State = item.State
	prop.Timeout = item.Timeout

	if len(item.Timestamp) == 0 {
		prop.LastUpdated = time.Now()
	} else {
		var err error
		prop.LastUpdated, err = time.ParseInLocation("2006-01-02T15:04:05.9", item.Timestamp, time.UTC)

		if err != nil {
			c.log.WithField("timestamp", item.Timestamp).WithError(err).Warn("error in time.ParseInLocation")
			prop.LastUpdated = time.Now()
		}
	}

	for _, val := range item.Switches {
		v, ok := prop.Values[val.Name]
		if !ok {
			continue
		}

		v.Value = SwitchState(strings.TrimSpace(string(val.Value)))

		prop.Values[val.Name] = v
	}

	if len(item.Message) > 0 {
		prop.Messages = append(prop.Messages, Message{
			Message:   item.Message,
			Timestamp: time.Now(),
		})
	}

	device.SwitchProperties[item.Name] = prop

	c.devices.Store(item.Device, device)
}

func (c *INDIClient) setTextVector(item *setTextVector) {
	device, err := c.findDevice(item.Device)
	if err != nil {
		c.log.WithField("device", item.Device).WithError(err).Warn("could not find device")
		return
	}

	var prop TextProperty
	if p, ok := device.TextProperties[item.Name]; ok {
		prop = p
	} else {
		c.log.WithField("device", item.Device).WithField("property", item.Name).Warn("could not find property")
		return
	}

	prop.State = item.State
	prop.Timeout = item.Timeout

	if len(item.Timestamp) == 0 {
		prop.LastUpdated = time.Now()
	} else {
		var err error
		prop.LastUpdated, err = time.ParseInLocation("2006-01-02T15:04:05.9", item.Timestamp, time.UTC)

		if err != nil {
			c.log.WithField("timestamp", item.Timestamp).WithError(err).Warn("error in time.ParseInLocation")
			prop.LastUpdated = time.Now()
		}
	}

	for _, val := range item.Texts {
		v, ok := prop.Values[val.Name]
		if !ok {
			continue
		}

		v.Value = strings.TrimSpace(val.Value)

		prop.Values[val.Name] = v
	}

	if len(item.Message) > 0 {
		prop.Messages = append(prop.Messages, Message{
			Message:   item.Message,
			Timestamp: time.Now(),
		})
	}

	device.TextProperties[item.Name] = prop

	c.devices.Store(item.Device, device)
}

func (c *INDIClient) setNumberVector(item *setNumberVector) {
	device, err := c.findDevice(item.Device)
	if err != nil {
		c.log.WithField("device", item.Device).WithError(err).Warn("could not find device")
		return
	}

	var prop NumberProperty
	if p, ok := device.NumberProperties[item.Name]; ok {
		prop = p
	} else {
		c.log.WithField("device", item.Device).WithField("property", item.Name).Warn("could not find property")
		return
	}

	prop.State = item.State
	prop.Timeout = item.Timeout

	if len(item.Timestamp) == 0 {
		prop.LastUpdated = time.Now()
	} else {
		var err error
		prop.LastUpdated, err = time.ParseInLocation("2006-01-02T15:04:05.9", item.Timestamp, time.UTC)

		if err != nil {
			c.log.WithField("timestamp", item.Timestamp).WithError(err).Warn("error in time.ParseInLocation")
			prop.LastUpdated = time.Now()
		}
	}

	for _, val := range item.Numbers {
		v, ok := prop.Values[val.Name]
		if !ok {
			continue
		}

		v.Value = strings.TrimSpace(val.Value)

		prop.Values[val.Name] = v
	}

	if len(item.Message) > 0 {
		prop.Messages = append(prop.Messages, Message{
			Message:   item.Message,
			Timestamp: time.Now(),
		})
	}

	device.NumberProperties[item.Name] = prop

	c.devices.Store(item.Device, device)
}

func (c *INDIClient) setLightVector(item *setLightVector) {
	device, err := c.findDevice(item.Device)
	if err != nil {
		c.log.WithField("device", item.Device).WithError(err).Warn("could not find device")
		return
	}

	var prop LightProperty
	if p, ok := device.LightProperties[item.Name]; ok {
		prop = p
	} else {
		c.log.WithField("device", item.Device).WithField("property", item.Name).Warn("could not find property")
		return
	}

	prop.State = item.State

	if len(item.Timestamp) == 0 {
		prop.LastUpdated = time.Now()
	} else {
		var err error
		prop.LastUpdated, err = time.ParseInLocation("2006-01-02T15:04:05.9", item.Timestamp, time.UTC)

		if err != nil {
			c.log.WithField("timestamp", item.Timestamp).WithError(err).Warn("error in time.ParseInLocation")
			prop.LastUpdated = time.Now()
		}
	}

	for _, val := range item.Lights {
		v, ok := prop.Values[val.Name]
		if !ok {
			continue
		}

		v.Value = PropertyState(strings.TrimSpace(string(val.Value)))

		prop.Values[val.Name] = v
	}

	if len(item.Message) > 0 {
		prop.Messages = append(prop.Messages, Message{
			Message:   item.Message,
			Timestamp: time.Now(),
		})
	}

	device.LightProperties[item.Name] = prop

	c.devices.Store(item.Device, device)
}

func (c *INDIClient) setBlobVector(item *setBlobVector) {
	device, err := c.findDevice(item.Device)
	if err != nil {
		c.log.WithField("device", item.Device).WithError(err).Warn("could not find device")
		return
	}

	var prop BlobProperty
	if p, ok := device.BlobProperties[item.Name]; ok {
		prop = p
	} else {
		c.log.WithField("device", item.Device).WithField("property", item.Name).Warn("could not find property")
		return
	}

	prop.State = item.State
	prop.Timeout = item.Timeout

	if len(item.Timestamp) == 0 {
		prop.LastUpdated = time.Now()
	} else {
		var err error
		prop.LastUpdated, err = time.ParseInLocation("2006-01-02T15:04:05.9", item.Timestamp, time.UTC)

		if err != nil {
			c.log.WithField("timestamp", item.Timestamp).WithError(err).Warn("error in time.ParseInLocation")
			prop.LastUpdated = time.Now()
		}
	}

	for _, val := range item.Blobs {
		v, ok := prop.Values[val.Name]
		if !ok {
			continue
		}

		fname := fmt.Sprintf("%s_%s_%s%s", item.Device, item.Name, val.Name, val.Format)

		f, err := c.fs.OpenFile(fname, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		if err != nil {
			c.log.WithField("file", fname).WithError(err).Warn("error in c.fs.OpenFile")
			continue
		}

		var writers []io.Writer

		if ws, ok := c.blobStreams.Load(fmt.Sprintf("%s_%s_%s", item.Device, item.Name, val.Name)); ok {
			wss := ws.(map[string]io.Writer)

			for _, w := range wss {
				writers = append(writers, w)
			}
		}

		writers = append(writers, f)

		val.Value = strings.TrimSpace(val.Value)
		r := base64.NewDecoder(base64.StdEncoding, strings.NewReader(val.Value))

		dest := io.MultiWriter(writers...)

		written, err := io.Copy(dest, r)
		if err != nil {
			c.log.WithError(err).Warn("error in io.Copy")
			continue
		}

		v.Value = f.Name()
		v.Size = written

		f.Close()

		prop.Values[val.Name] = v
	}

	if len(item.Message) > 0 {
		prop.Messages = append(prop.Messages, Message{
			Message:   item.Message,
			Timestamp: time.Now(),
		})
	}

	device.BlobProperties[item.Name] = prop

	c.devices.Store(item.Device, device)
}

func (c *INDIClient) message(item *message) {
	device, err := c.findDevice(item.Device)
	if err != nil {
		c.log.WithField("device", item.Device).WithError(err).Warn("could not find device")
		return
	}

	device.Messages = append(device.Messages, Message{
		Message:   item.Message,
		Timestamp: time.Now(),
	})

	c.devices.Store(item.Device, device)
}

func (c *INDIClient) delProperty(item *delProperty) {
	if len(item.Device) == 0 {
		c.devices.Range(func(key, value interface{}) bool {
			c.devices.Delete(key)
			return true
		})

		return
	}

	if len(item.Name) == 0 {
		c.devices.Delete(item.Device)
		return
	}

	device := c.findOrCreateDevice(item.Device)

	delete(device.TextProperties, item.Name)
	delete(device.NumberProperties, item.Name)
	delete(device.SwitchProperties, item.Name)
	delete(device.LightProperties, item.Name)
	delete(device.BlobProperties, item.Name)

	c.devices.Store(item.Device, device)
}

func (c *INDIClient) startRead() {
	go func(r <-chan interface{}, log logging.Logger, handler indiMessageHandler) {
		for i := range r {
			log.WithField("item", i).Debug("got message")

			switch item := i.(type) {
			case *defTextVector:
				handler.defTextVector(item)
			case *defSwitchVector:
				handler.defSwitchVector(item)
			case *defNumberVector:
				handler.defNumberVector(item)
			case *defLightVector:
				handler.defLightVector(item)
			case *defBlobVector:
				handler.defBlobVector(item)
			case *setSwitchVector:
				handler.setSwitchVector(item)
			case *setTextVector:
				handler.setTextVector(item)
			case *setNumberVector:
				handler.setNumberVector(item)
			case *setLightVector:
				handler.setLightVector(item)
			case *setBlobVector:
				handler.setBlobVector(item)
			case *message:
				handler.message(item)
			case *delProperty:
				handler.delProperty(item)
			default:
				log.WithField("type", fmt.Sprintf("%T", item)).Warn("unknown type")
			}
		}
	}(c.read, c.log, c)

	go func(conn io.Reader, r chan<- interface{}, log logging.Logger) {
		decoder := xml.NewDecoder(conn)

		var inElement string
		for {
			t, err := decoder.Token()
			if err != nil {
				log.WithError(err).Warn("error in decoder.Token")

				if err == io.EOF {
					c.Disconnect()
					return
				}
				continue
			}

			var item interface{}

			switch se := t.(type) {
			case xml.StartElement:
				log.WithField("startElement", se.Name.Local).Debug("read start element")

				var inner interface{}
				inElement = se.Name.Local
				switch inElement {
				case "defSwitchVector":
					inner = &defSwitchVector{}
				case "defTextVector":
					inner = &defTextVector{}
				case "defNumberVector":
					inner = &defNumberVector{}
				case "defLightVector":
					inner = &defLightVector{}
				case "defBLOBVector":
					inner = &defBlobVector{}
				case "setSwitchVector":
					inner = &setSwitchVector{}
				case "setTextVector":
					inner = &setTextVector{}
				case "setNumberVector":
					inner = &setNumberVector{}
				case "setLightVector":
					inner = &setLightVector{}
				case "setBLOBVector":
					inner = &setBlobVector{}
				case "message":
					inner = &message{}
				case "delProperty":
					inner = &delProperty{}
				default:
					log.WithField("element", inElement).Error("unknown element")
				}

				if inner != nil {
					err = decoder.DecodeElement(&inner, &se)
					if err != nil {
						log.WithField("element", inElement).WithError(err).Error("error in decoder.DecodeElement")
						continue
					}

					item = inner
				}
			}

			if item != nil {
				r <- item
			}
		}
	}(c.conn, c.read, c.log)
}

func (c *INDIClient) startWrite() {
	go func(conn io.Writer, w <-chan interface{}, log logging.Logger) {
		for item := range w {
			b, err := xml.Marshal(item)
			if err != nil {
				log.WithError(err).Error("error in xml.Marshal")
				continue
			}

			log.WithField("cmd", string(b)).Debug("sending command")

			_, err = conn.Write(b)
			if err != nil {
				log.WithError(err).Error("error in conn.Write")
				continue
			}
		}
	}(c.conn, c.write, c.log)
}
