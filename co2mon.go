package co2_monitor

import (
	"context"
	"errors"
	"fmt"
	"github.com/sstallion/go-hid"
	"strconv"
	"time"
)

const (
	vendorID  uint16 = 0x04d9
	productID uint16 = 0xa052

	packageSize = 8

	co2Code  = 0x50
	tempCode = 0x42

	minCO2 = 0
	maxCO2 = 3000

	absZeroTemp = 273.15
	tempFactor  = 0.0625
)

func Start(ctx context.Context, delay time.Duration) (chan Metric, error) {
	if err := hid.Init(); err != nil {
		return nil, err
	}

	device, err := hid.OpenFirst(vendorID, productID)
	if err != nil {
		return nil, err
	}

	magicTable := make([]byte, packageSize+1)
	res, err := device.SendFeatureReport(magicTable)
	if err != nil {
		return nil, err
	}
	if res != packageSize+1 {
		return nil, errors.New("unable to send magic table to CO2 device")
	}

	metricChan := make(chan Metric)
	ticker := time.NewTicker(delay)
	go func() {
		for {
			select {
			case <-ctx.Done():
				device.Close()
				hid.Exit()
				ticker.Stop()
				close(metricChan)
				return
			case <-ticker.C:
				metricChan <- readMetric(device)
			}
		}
	}()

	return metricChan, nil
}

func readMetric(device *hid.Device) Metric {
	for {
		data := make([]byte, packageSize)
		_, err := device.ReadWithTimeout(data, 5*time.Second)
		if err != nil {
			return Metric{Type: Error, Time: time.Now(), Value: err.Error()}
		}

		if !checkCRC(data) {
			return Metric{Type: Error, Time: time.Now(), Value: "bad crc"}
		}

		metric := parseMetric(data)
		if metric.Type == Unknown {
			continue
		}

		return metric
	}
}

func checkCRC(data []byte) bool {
	if len(data) < 5 {
		return false
	}

	return data[4] == 0x0d && data[0]+data[1]+data[2] == data[3]
}

func parseMetric(data []byte) Metric {
	if len(data) < 5 {
		return Metric{
			Type:  Error,
			Time:  time.Now(),
			Value: fmt.Sprintf("wrong data length %d but must be >= 5", len(data)),
		}
	}

	value := int(data[1])<<8 | int(data[2])
	if data[0] == co2Code {
		if value < minCO2 {
			return Metric{
				Type:  Error,
				Time:  time.Now(),
				Value: fmt.Sprintf("co2 ppm is less than min value %d", minCO2),
			}
		}
		if value > maxCO2 {
			value = maxCO2
		}
		return Metric{
			Type:  CO2,
			Value: strconv.Itoa(value),
			Time:  time.Now(),
		}
	}

	if data[0] == tempCode {
		return Metric{
			Type:  Temp,
			Value: strconv.FormatFloat(float64(value)*tempFactor-absZeroTemp, 'f', 1, 64),
			Time:  time.Now(),
		}
	}

	return Metric{Type: Unknown}
}
