package config

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadCalibrationFile reads a normalized calibration artifact from YAML, JSON, or CSV.
func LoadCalibrationFile(filename string) (*CalibrationConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read calibration file %s: %w", filename, err)
	}

	switch strings.ToLower(filepath.Ext(filename)) {
	case ".yaml", ".yml":
		return parseCalibrationYAML(data)
	case ".json":
		return parseCalibrationJSON(data)
	case ".csv":
		return parseCalibrationCSV(data)
	default:
		return nil, fmt.Errorf("unsupported calibration file type %q", filepath.Ext(filename))
	}
}

func parseCalibrationYAML(data []byte) (*CalibrationConfig, error) {
	var wrapper struct {
		Calibration *CalibrationConfig `yaml:"calibration"`
	}
	if err := yaml.Unmarshal(data, &wrapper); err == nil && wrapper.Calibration != nil {
		return wrapper.Calibration, nil
	}

	var calibration CalibrationConfig
	if err := yaml.Unmarshal(data, &calibration); err != nil {
		return nil, fmt.Errorf("failed to parse calibration YAML: %w", err)
	}
	return &calibration, nil
}

func parseCalibrationJSON(data []byte) (*CalibrationConfig, error) {
	var wrapper struct {
		Calibration *CalibrationConfig `json:"calibration"`
	}
	if err := json.Unmarshal(data, &wrapper); err == nil && wrapper.Calibration != nil {
		return wrapper.Calibration, nil
	}

	var calibration CalibrationConfig
	if err := json.Unmarshal(data, &calibration); err != nil {
		return nil, fmt.Errorf("failed to parse calibration JSON: %w", err)
	}
	return &calibration, nil
}

func parseCalibrationCSV(data []byte) (*CalibrationConfig, error) {
	reader := csv.NewReader(strings.NewReader(string(data)))
	reader.TrimLeadingSpace = true

	calibration := &CalibrationConfig{
		PlanMix: map[string]float64{},
		Usage:   map[string]float64{},
	}

	rowNumber := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to parse calibration CSV: %w", err)
		}
		rowNumber++
		if len(record) < 2 {
			continue
		}

		key := strings.TrimSpace(record[0])
		value := strings.TrimSpace(record[1])
		if key == "" {
			continue
		}
		if rowNumber == 1 && strings.EqualFold(key, "field") && strings.EqualFold(value, "value") {
			continue
		}

		if err := applyCalibrationCSVField(calibration, key, value); err != nil {
			return nil, fmt.Errorf("invalid calibration CSV row %d (%s): %w", rowNumber, key, err)
		}
	}

	return calibration, nil
}

func applyCalibrationCSVField(calibration *CalibrationConfig, key, rawValue string) error {
	switch key {
	case "period":
		calibration.Period = rawValue
		return nil
	case "source":
		calibration.Source = rawValue
		return nil
	case "gross_revenue":
		value, err := strconv.ParseFloat(rawValue, 64)
		if err != nil {
			return err
		}
		calibration.GrossRevenue = value
		return nil
	case "payment_fees":
		value, err := strconv.ParseFloat(rawValue, 64)
		if err != nil {
			return err
		}
		calibration.PaymentFees = value
		return nil
	case "free_users":
		value, err := strconv.Atoi(rawValue)
		if err != nil {
			return err
		}
		calibration.FreeUsers = value
		return nil
	case "paid_users.monthly":
		value, err := strconv.Atoi(rawValue)
		if err != nil {
			return err
		}
		if calibration.PaidUsers == nil {
			calibration.PaidUsers = &CalibrationPaidUsers{}
		}
		calibration.PaidUsers.Monthly = value
		return nil
	case "paid_users.annual":
		value, err := strconv.Atoi(rawValue)
		if err != nil {
			return err
		}
		if calibration.PaidUsers == nil {
			calibration.PaidUsers = &CalibrationPaidUsers{}
		}
		calibration.PaidUsers.Annual = value
		return nil
	default:
		switch {
		case strings.HasPrefix(key, "plan_mix."):
			value, err := strconv.ParseFloat(rawValue, 64)
			if err != nil {
				return err
			}
			calibration.PlanMix[strings.TrimPrefix(key, "plan_mix.")] = value
			return nil
		case strings.HasPrefix(key, "usage."):
			value, err := strconv.ParseFloat(rawValue, 64)
			if err != nil {
				return err
			}
			calibration.Usage[strings.TrimPrefix(key, "usage.")] = value
			return nil
		default:
			return fmt.Errorf("unsupported field")
		}
	}
}
