package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type GPIOLine struct {
	Number uint
	fd     *os.File
}

const (
	IN = iota
	OUT
)

func NewGPIOLine(number uint, direction int) (gpio *GPIOLine, err error) {
	gpio = new(GPIOLine)
	gpio.Number = number

	if err := gpio.enable_export(); err != nil {
		return nil, err
	}
	err = gpio.SetDirection(direction)
	if err != nil {
		return nil, err
	}
	gpio.fd, err = os.OpenFile(fmt.Sprintf("/sys/class/gpio/gpio%d/value", gpio.Number), os.O_WRONLY|os.O_SYNC, 0666)
	if err != nil {
		return nil, err
	}
	return gpio, nil
}

func (gpio *GPIOLine) enable_export() error {
	_, err := os.Stat(fmt.Sprintf("/sys/class/gpio/gpio%d", gpio.Number))
	if err == nil {
		// already exported
		return nil
	} else if err != nil && !os.IsNotExist(err) {
		// some other error
		return err
	}
	fd, err := os.OpenFile("/sys/class/gpio/export", os.O_WRONLY|os.O_SYNC, 0666)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(fd, "%d\n", gpio.Number)
	return err
}

func (gpio *GPIOLine) SetDirection(direction int) error {
	df, err := os.OpenFile(fmt.Sprintf("/sys/class/gpio/gpio%d/direction", gpio.Number),
		os.O_WRONLY|os.O_SYNC, 0666)
	if err != nil {
		return err
	}
	fmt.Fprintln(df, "out")
	df.Close()
	return nil
}

func (gpio *GPIOLine) SetState(state bool) error {
	v := "0"
	if state {
		v = "1"
	}
	_, err := fmt.Fprintln(gpio.fd, v)
	return err
}

func (gpio *GPIOLine) Close() {
	gpio.fd.Close()
}

type DS18S20 struct {
	Id         uint64 // Actual 48-bit ID burned into chip
	Name       string // Name assigned by linux w1 drivers
	FamilyCode uint8  // Manufacturer-assigned family code
	fd         *os.File
	reader     *bufio.Reader
}

const (
	MODEL_DS18S20 = 0x10
	MODEL_DS18B20 = 0x28
)

func (d *DS18S20) Model() string {
	switch {
	case d.FamilyCode == MODEL_DS18S20:
		return "ds18s20"
	case d.FamilyCode == MODEL_DS18B20:
		return "ds18b20"
	}
	// Shouldn't get here, since we checked back in NewDS18S20
	return ""
}

func (d *DS18S20) HumanId() string {
	return fmt.Sprintf("%s-%012x", d.Model(), d.Id)
}

func NewDS18S20(name string) (*DS18S20, error) {
	device := new(DS18S20)
	device.Name = name

	var err error
	if device.FamilyCode, device.Id, err = read_device_id(device); err != nil {
		return nil, err
	}
	if device.Model() == "" {
		return nil, fmt.Errorf("Unrecognized/unsupported 1-wire family code 0x%x", device.FamilyCode)
	}
	device.fd, err = os.OpenFile(fmt.Sprintf("/sys/bus/w1/devices/%v/w1_slave", device.Name), os.O_RDONLY|os.O_SYNC, 0666)
	if err != nil {
		return nil, err
	}
	device.reader = bufio.NewReader(device.fd)
	return device, nil
}

func read_device_id(device *DS18S20) (uint8, uint64, error) {
	fn := fmt.Sprintf("/sys/bus/w1/devices/%v/id", device.Name)
	id_fd, err := os.OpenFile(fn, os.O_RDONLY, 0666)
	if err != nil {
		return 0, 0, err
	}
	defer id_fd.Close()
	var romcode uint64
	err = binary.Read(id_fd, binary.LittleEndian, &romcode)
	if err != nil {
		return 0, 0, fmt.Errorf("Error decoding %v device id: %v", fn, err)
	}
	devicetype := uint8(romcode & 0xff)
	id := (romcode & 0x00ffffffffffff00) >> 8
	return devicetype, id, nil
}

var __CRC_CHECK_REGEX *regexp.Regexp = regexp.MustCompile(`crc=\w+\s(YES|NO)`)
var __TEMP_SAMPLE_REGEX *regexp.Regexp = regexp.MustCompile(`.*\st=(\d+)`)

func (device *DS18S20) Read() (millidegrees int64, err error) {
	device.fd.Seek(0, 0)
	for {
		line, err := device.reader.ReadString('\n')
		if err != nil {
			return 0., err
		} else if len(line) == 0 {
			return 0., fmt.Errorf("EOF without data from w1")
		} else {
			// fmt.Printf("read from %v: %v", device.Id, line)
			matches := __CRC_CHECK_REGEX.FindStringSubmatch(line)
			if len(matches) > 0 && matches[1] != "YES" {
				return 0., fmt.Errorf("CRC mismatch on read")
			}

			matches = __TEMP_SAMPLE_REGEX.FindStringSubmatch(line)
			if len(matches) > 0 {
				// fmt.Printf("matched for temp: %v", matches)
				v, err := strconv.ParseInt(matches[1], 10, 64)
				if err != nil {
					return 0., err
				}
				return v, nil
			}
		}
	}
	// return 0., fmt.Errorf("Malformed/empty output from w1 slave")
}

/* Find all attached 1-wire slave devices */
func ScanSlaves() ([]string, error) {
	devicedir, err := os.Open("/sys/bus/w1/devices/")
	if err != nil {
		return nil, err
	}
	defer devicedir.Close()
	names, err := devicedir.Readdirnames(0)
	if err != nil {
		return nil, err
	}
	r := make([]string, 0, len(names)-1)
	for i := range names {
		if !strings.Contains(names[i], "w1_bus_master") {
			r = append(r, names[i])
		}
	}
	return r, nil
}
