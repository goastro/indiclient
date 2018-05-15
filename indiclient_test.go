package indiclient

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/rickbassham/logging"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockDialer struct {
	mock.Mock
}

func (m *mockDialer) Dial(network, address string) (io.ReadWriteCloser, error) {
	args := m.Called(network, address)

	c := args.Get(0)
	err := args.Error(1)
	if c == nil {
		return nil, err
	}

	return c.(io.ReadWriteCloser), err
}

type mockConnection struct {
	w *bytes.Buffer
	r *bytes.Buffer
}

func (m *mockConnection) Read(p []byte) (n int, err error) {
	n, err = m.r.Read(p)

	if err == io.EOF {
		for {
			// simulate no data ready on an open connection
			time.Sleep(1 * time.Second)
		}
	}

	return
}

func (m *mockConnection) Write(p []byte) (n int, err error) {
	return m.w.Write(p)
}

func (m *mockConnection) Close() error {
	return nil
}

func TestClient(t *testing.T) {
	testXML := `<defSwitchVector device="Camera" name="Binning" rule="OneOfMany" state="Ok" perm="w" timeout="0"
	label="Binning">
   <defSwitch name="One" label="1:1">Off</defSwitch>
   <defSwitch name="Two" label="2:1">On </defSwitch>
   <defSwitch name="Three" label="3:1">Off</defSwitch>
   <defSwitch name="Four" label="4:1">Off</defSwitch>
   </defSwitchVector>`

	r := bytes.NewBufferString(testXML)

	written := []byte{}

	w := bytes.NewBuffer(written)

	conn := &mockConnection{
		r: r,
		w: w,
	}

	network := "tcp"
	address := "localhost:1"

	dialer := &mockDialer{}
	dialer.On("Dial", network, address).Return(conn, nil)

	log := logging.NewLogger(os.Stdout, logging.JSONFormatter{}, logging.LogLevelInfo)
	fs := afero.NewMemMapFs()

	c := NewINDIClient(log, dialer, fs, 5)

	err := c.Connect(network, address)
	require.NoError(t, err)

	time.Sleep(1 * time.Second) // Wait for the client to read the xml

	devices := c.Devices()
	require.Len(t, devices, 1)

	err = c.Disconnect()
	require.NoError(t, err)
}

func Test_DialerError(t *testing.T) {
	network := "tcp"
	address := "localhost:1"

	dialer := &mockDialer{}
	dialer.On("Dial", network, address).Return(nil, errors.New("some error"))

	log := logging.NewLogger(os.Stdout, logging.JSONFormatter{}, logging.LogLevelInfo)
	fs := afero.NewMemMapFs()

	c := NewINDIClient(log, dialer, fs, 5)

	err := c.Connect(network, address)
	require.Error(t, err)
	assert.EqualError(t, err, "some error")
}

func Test_DisconnectWithoutConnect(t *testing.T) {
	log := logging.NewLogger(os.Stdout, logging.JSONFormatter{}, logging.LogLevelInfo)
	fs := afero.NewMemMapFs()

	dialer := &mockDialer{}
	c := NewINDIClient(log, dialer, fs, 5)

	err := c.Disconnect()
	assert.NoError(t, err)
}

func Test_GetProperties(t *testing.T) {
	r := bytes.NewBufferString("")

	written := []byte{}

	w := bytes.NewBuffer(written)

	conn := &mockConnection{
		r: r,
		w: w,
	}

	network := "tcp"
	address := "localhost:1"

	dialer := &mockDialer{}
	dialer.On("Dial", network, address).Return(conn, nil)

	log := logging.NewLogger(os.Stdout, logging.JSONFormatter{}, logging.LogLevelInfo)
	fs := afero.NewMemMapFs()

	c := NewINDIClient(log, dialer, fs, 5)

	err := c.Connect(network, address)
	require.NoError(t, err)

	err = c.GetProperties("", "")
	require.NoError(t, err)

	time.Sleep(1 * time.Second) // Wait for the client to write the xml

	result := w.String()

	assert.Equal(t, "<getProperties version=\"1.7\"></getProperties>", result)

	err = c.Disconnect()
	require.NoError(t, err)
}

func Test_GetProperties_PropWithNoDevice(t *testing.T) {
	r := bytes.NewBufferString("")

	written := []byte{}

	w := bytes.NewBuffer(written)

	conn := &mockConnection{
		r: r,
		w: w,
	}

	network := "tcp"
	address := "localhost:1"

	dialer := &mockDialer{}
	dialer.On("Dial", network, address).Return(conn, nil)

	log := logging.NewLogger(os.Stdout, logging.JSONFormatter{}, logging.LogLevelInfo)
	fs := afero.NewMemMapFs()

	c := NewINDIClient(log, dialer, fs, 5)

	err := c.Connect(network, address)
	require.NoError(t, err)

	err = c.GetProperties("", "prop1")
	require.Error(t, err)
	assert.EqualError(t, err, ErrPropertyWithoutDevice.Error())

	err = c.Disconnect()
	require.NoError(t, err)
}

func Test_EnableBlob_MissingDevice(t *testing.T) {
	r := bytes.NewBufferString("")

	written := []byte{}

	w := bytes.NewBuffer(written)

	conn := &mockConnection{
		r: r,
		w: w,
	}

	c := &INDIClient{
		devices: sync.Map{},
		conn:    conn,
	}

	err := c.EnableBlob("", "", BlobEnableAlso)
	require.Error(t, err)
	assert.EqualError(t, err, ErrDeviceNotFound.Error())

	err = c.Disconnect()
	require.NoError(t, err)
}

func Test_EnableBlob_InvalidValue(t *testing.T) {
	r := bytes.NewBufferString("")

	written := []byte{}

	w := bytes.NewBuffer(written)

	conn := &mockConnection{
		r: r,
		w: w,
	}

	c := &INDIClient{
		devices: sync.Map{},
		conn:    conn,
	}

	c.devices.Store("device1", Device{
		Name: "device1",
	})

	err := c.EnableBlob("device1", "", BlobEnable("test"))
	require.Error(t, err)
	assert.EqualError(t, err, ErrInvalidBlobEnable.Error())

	err = c.Disconnect()
	require.NoError(t, err)
}

func Test_EnableBlob_Success(t *testing.T) {
	r := bytes.NewBufferString("")

	written := []byte{}

	w := bytes.NewBuffer(written)

	conn := &mockConnection{
		r: r,
		w: w,
	}

	network := "tcp"
	address := "localhost:1"

	dialer := &mockDialer{}
	dialer.On("Dial", network, address).Return(conn, nil)

	log := logging.NewLogger(os.Stdout, logging.JSONFormatter{}, logging.LogLevelInfo)
	fs := afero.NewMemMapFs()

	c := NewINDIClient(log, dialer, fs, 5)

	err := c.Connect(network, address)
	require.NoError(t, err)

	c.devices.Store("device1", Device{
		Name: "device1",
	})

	err = c.EnableBlob("device1", "", BlobEnableAlso)
	require.NoError(t, err)

	time.Sleep(1 * time.Second) // Wait for the client to write the xml

	result := w.String()

	assert.Equal(t, "<enableBLOB device=\"device1\" name=\"\">Also</enableBLOB>", result)

	err = c.Disconnect()
	require.NoError(t, err)
}

func TestModels_defSwitchVector(t *testing.T) {
	testCases := []struct {
		xml      string
		bind     defSwitchVector
		expected defSwitchVector
	}{
		{
			xml: `<defSwitchVector device="Camera" name="Binning" rule="OneOfMany" state="Ok" perm="wo" timeout="0"
		label="Binning">
	   <defSwitch name="One" label="1:1">Off</defSwitch>
	   <defSwitch name="Two" label="2:1">On</defSwitch>
	   <defSwitch name="Three" label="3:1">Off</defSwitch>
	   <defSwitch name="Four" label="4:1">Off</defSwitch>
	   </defSwitchVector>`,
			bind: defSwitchVector{},
			expected: defSwitchVector{
				XMLName: xml.Name{
					Local: "defSwitchVector",
				},
				Device:  "Camera",
				Name:    "Binning",
				Rule:    SwitchRuleOneOfMany,
				State:   PropertyStateOk,
				Perm:    PropertyPermissionWriteOnly,
				Timeout: 0,
				Label:   "Binning",
				Switches: []defSwitch{
					defSwitch{
						XMLName: xml.Name{
							Local: "defSwitch",
						},
						Name:  "One",
						Label: "1:1",
						Value: "Off",
					},
					defSwitch{
						XMLName: xml.Name{
							Local: "defSwitch",
						},
						Name:  "Two",
						Label: "2:1",
						Value: "On",
					},
					defSwitch{
						XMLName: xml.Name{
							Local: "defSwitch",
						},
						Name:  "Three",
						Label: "3:1",
						Value: "Off",
					},
					defSwitch{
						XMLName: xml.Name{
							Local: "defSwitch",
						},
						Name:  "Four",
						Label: "4:1",
						Value: "Off",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		err := xml.Unmarshal([]byte(tc.xml), &tc.bind)
		require.NoError(t, err)

		assert.Equal(t, tc.expected, tc.bind)
	}
}

func Test_GetBlob_MissingDevice(t *testing.T) {
	c := &INDIClient{
		devices: sync.Map{},
	}

	rdr, name, size, err := c.GetBlob("device1", "prop1", "blob1")

	require.Nil(t, rdr)
	require.Empty(t, name)
	require.Zero(t, size)
	require.Error(t, err)
	assert.EqualError(t, err, ErrDeviceNotFound.Error())
}

func Test_GetBlob_MissingProperty(t *testing.T) {
	c := &INDIClient{
		devices: sync.Map{},
	}

	c.devices.Store("device1", Device{
		Name: "device1",
	})

	rdr, name, size, err := c.GetBlob("device1", "prop1", "blob1")

	require.Nil(t, rdr)
	require.Empty(t, name)
	require.Zero(t, size)
	require.Error(t, err)
	assert.EqualError(t, err, ErrPropertyNotFound.Error())
}

func Test_GetBlob_MissingValue(t *testing.T) {
	c := &INDIClient{
		devices: sync.Map{},
	}

	c.devices.Store("device1", Device{
		Name: "device1",
		BlobProperties: map[string]BlobProperty{
			"prop1": BlobProperty{
				Name: "prop1",
			},
		},
	})

	rdr, name, size, err := c.GetBlob("device1", "prop1", "blob1")

	require.Nil(t, rdr)
	require.Empty(t, name)
	require.Zero(t, size)
	require.Error(t, err)
	assert.EqualError(t, err, ErrPropertyValueNotFound.Error())
}

func Test_GetBlob_FileNotFound(t *testing.T) {
	c := &INDIClient{
		devices: sync.Map{},
		fs:      afero.NewMemMapFs(),
	}

	c.devices.Store("device1", Device{
		Name: "device1",
		BlobProperties: map[string]BlobProperty{
			"prop1": BlobProperty{
				Name: "prop1",
				Values: map[string]BlobValue{
					"blob1": BlobValue{
						Name:  "blob1",
						Value: "file.fit",
						Size:  13,
						Label: "label1",
					},
				},
			},
		},
	})

	rdr, name, size, err := c.GetBlob("device1", "prop1", "blob1")

	require.Nil(t, rdr)
	require.Empty(t, name)
	require.Zero(t, size)
	require.Error(t, err)
	assert.Equal(t, err, &os.PathError{Op: "open", Path: "file.fit", Err: afero.ErrFileNotFound})
}

func Test_GetBlob_Success(t *testing.T) {
	c := &INDIClient{
		devices: sync.Map{},
		fs:      afero.NewMemMapFs(),
	}

	f, _ := c.fs.Create("file.fit")
	f.WriteString("1234567890")

	c.devices.Store("device1", Device{
		Name: "device1",
		BlobProperties: map[string]BlobProperty{
			"prop1": BlobProperty{
				Name: "prop1",
				Values: map[string]BlobValue{
					"blob1": BlobValue{
						Name:  "blob1",
						Value: "file.fit",
						Size:  10,
						Label: "label1",
					},
				},
			},
		},
	})

	rdr, name, size, err := c.GetBlob("device1", "prop1", "blob1")

	require.NoError(t, err)
	require.NotNil(t, rdr)
	assert.Equal(t, name, "file.fit")
	assert.Equal(t, int64(10), size)

	b, _ := ioutil.ReadAll(rdr)
	assert.Equal(t, "1234567890", string(b))
}

func Example_singleClient() {
	var err error

	log := logging.NewLogger(os.Stdout, logging.JSONFormatter{}, logging.LogLevelInfo)
	dialer := NetworkDialer{}
	fs := afero.NewMemMapFs()
	bufferSize := 10

	// Initialize a new INDIClient.
	client := NewINDIClient(log, dialer, fs, bufferSize)

	// Connect to the local indiserver.
	err = client.Connect("tcp", "localhost:7624")
	if err != nil {
		panic(err.Error())
	}

	// Get all properties of all devices.
	err = client.GetProperties("", "")
	if err != nil {
		panic(err.Error())
	}

	// Wait to get the devices back from indiserver.
	time.Sleep(2 * time.Second)

	// Print the names of all the devices we found.
	devices := client.Devices()
	for _, device := range devices {
		println(device.Name)
	}

	// Connect to our ASI224MC camera.
	err = client.SetSwitchValue("ZWO CCD ASI224MC", "CONNECTION", "CONNECT", SwitchStateOn)
	if err != nil {
		panic(err.Error())
	}

	// Tell the indiserver we want blobs from this camera's CCD1 property.
	err = client.EnableBlob("ZWO CCD ASI224MC", "CCD1", BlobEnableAlso)
	if err != nil {
		panic(err.Error())
	}

	// Take a 10 second exposure.
	err = client.SetNumberValue("ZWO CCD ASI224MC", "CCD_EXPOSURE", "CCD_EXPOSURE_VALUE", "10")
	if err != nil {
		panic(err.Error())
	}

	// Wait for the exposure to finish and transfer.
	time.Sleep(11 * time.Second)

	// Get the actual BLOB. Be sure to close rdr when you are done with it!
	rdr, fileName, length, err := client.GetBlob("ZWO CCD ASI224MC", "CCD1", "CCD1")
	if err != nil {
		panic(err.Error())
	}

	println(fmt.Sprintf("%s %d", fileName, length))

	err = rdr.Close()
	if err != nil {
		panic(err.Error())
	}
}

func Example_multipleClients() {
	var err error

	log := logging.NewLogger(os.Stdout, logging.JSONFormatter{}, logging.LogLevelInfo)
	dialer := NetworkDialer{}
	fs := afero.NewMemMapFs()
	bufferSize := 10
	blobfs := afero.NewMemMapFs()

	// Initialize a new INDIClient.
	client := NewINDIClient(log, dialer, fs, bufferSize)

	// Connect to the local indiserver.
	err = client.Connect("tcp", "localhost:7624")
	if err != nil {
		panic(err.Error())
	}

	// Get all properties of all devices.
	err = client.GetProperties("", "")
	if err != nil {
		panic(err.Error())
	}

	// Wait to get the devices back from indiserver.
	time.Sleep(2 * time.Second)

	// Print the names of all the devices we found.
	devices := client.Devices()
	for _, device := range devices {
		println(device.Name)
	}

	// Connect to our ASI224MC camera.
	err = client.SetSwitchValue("ZWO CCD ASI224MC", "CONNECTION", "CONNECT", SwitchStateOn)
	if err != nil {
		panic(err.Error())
	}

	blobClient := NewINDIClient(log, dialer, blobfs, bufferSize)

	// Connect to the local indiserver.
	err = blobClient.Connect("tcp", "localhost:7624")
	if err != nil {
		panic(err.Error())
	}

	// Get the "CCD1" property of the "ZWO CCD ASI224MC" device.
	err = blobClient.GetProperties("ZWO CCD ASI224MC", "CCD1")
	if err != nil {
		panic(err.Error())
	}

	// Wait to get the devices back from indiserver.
	time.Sleep(2 * time.Second)

	// Tell the indiserver we want blobs from this camera's CCD1 property, and ONLY blobs from this camera.
	// This allows the other client to stay open for control data, without slowing things down with large
	// file transfers.
	err = blobClient.EnableBlob("ZWO CCD ASI224MC", "CCD1", BlobEnableOnly)
	if err != nil {
		panic(err.Error())
	}

	// Take a 10 second exposure. We send this on the control client.
	err = client.SetNumberValue("ZWO CCD ASI224MC", "CCD_EXPOSURE", "CCD_EXPOSURE_VALUE", "10")
	if err != nil {
		panic(err.Error())
	}

	// Wait for the exposure to finish and transfer.
	time.Sleep(11 * time.Second)

	// Get the actual BLOB. Be sure to close rdr when you are done with it!
	rdr, fileName, length, err := blobClient.GetBlob("ZWO CCD ASI224MC", "CCD1", "CCD1")
	if err != nil {
		panic(err.Error())
	}

	println(fmt.Sprintf("%s %d", fileName, length))

	err = rdr.Close()
	if err != nil {
		panic(err.Error())
	}

	err = client.Disconnect()
	if err != nil {
		panic(err.Error())
	}

	err = blobClient.Disconnect()
	if err != nil {
		panic(err.Error())
	}
}

func ExampleINDIClient_SetSwitchValue_connect() {
	var err error

	log := logging.NewLogger(os.Stdout, logging.JSONFormatter{}, logging.LogLevelInfo)
	dialer := NetworkDialer{}
	fs := afero.NewMemMapFs()
	bufferSize := 10

	// Initialize a new INDIClient.
	client := NewINDIClient(log, dialer, fs, bufferSize)

	// Connect to the local indiserver.
	err = client.Connect("tcp", "localhost:7624")
	if err != nil {
		panic(err.Error())
	}

	// Get all properties of all devices.
	err = client.GetProperties("", "")
	if err != nil {
		panic(err.Error())
	}

	// Wait to get the devices back from indiserver.
	time.Sleep(2 * time.Second)

	// Connect to our ASI224MC camera.
	err = client.SetSwitchValue("ZWO CCD ASI224MC", "CONNECTION", "CONNECT", SwitchStateOn)
	if err != nil {
		panic(err.Error())
	}

	// Wait to connect to the device.
	time.Sleep(2 * time.Second)

	// Notice that we are not setting "CONNECT" to SwitchStateOff, but instead setting "DISCONNECT" to SwitchStateOn.
	err = client.SetSwitchValue("ZWO CCD ASI224MC", "CONNECTION", "DISCONNECT", SwitchStateOn)
	if err != nil {
		panic(err.Error())
	}

	err = client.Disconnect()
	if err != nil {
		panic(err.Error())
	}
}
