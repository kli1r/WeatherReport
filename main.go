package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"github.com/pelletier/go-toml"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Weather struct {
	XMLName        xml.Name   `xml:"weather_report" json:"-" yaml:"-" toml:"-"`
	Type           string     `toml:"weather_type" json:"weather_type" xml:"weather_type" yaml:"weather_type"`
	LocationName   string     `toml:"location" json:"location" xml:"location" yaml:"location"`
	LocationCoords [2]float64 `toml:"location_coords" json:"location_coords" xml:"-" yaml:"location_coords,flow"`
	XCoord         float64    `xml:"location_coords>lat" toml:"-" json:"-" yaml:"-"`
	YCoord         float64    `xml:"location_coords>long" toml:"-" json:"-" yaml:"-"`
	Temperature    float64    `toml:"temperature" json:"temperature" xml:"temperature" yaml:"temperature"`
	Date           time.Time  `toml:"date" json:"date" xml:"date" yaml:"date"`
	Comment        string     `toml:"info" comment:"Info from observer" json:"info,omitempty" xml:"info,omitempty" yaml:"info,omitempty"`
}

var weather Weather
var snowWeather = flag.NewFlagSet("snowy", flag.ExitOnError)
var hotlyWeather = flag.NewFlagSet("hotly", flag.ExitOnError)

func parseLocationCoords(args string) error {
	coordinates := strings.Split(args, ",")
	if len(coordinates) != 2 {
		return fmt.Errorf("invalid coordinates number, must be 2")
	}

	for i, val := range coordinates {
		var err error
		weather.LocationCoords[i], err = strconv.ParseFloat(val, 32)
		if err != nil {
			return fmt.Errorf("These are not coordinates! Either their format is not float64")
		}
	}
	if weather.LocationCoords[0] < -90 || weather.LocationCoords[0] > 90 || weather.LocationCoords[1] < -180 || weather.LocationCoords[1] > 180 {
		return fmt.Errorf("The latitude should be in the range [-90, 90], and the longitude in [-180, 180]!")
	}

	weather.XCoord = weather.LocationCoords[0]
	weather.YCoord = weather.LocationCoords[1]
	return nil
}

func marshalIndentJSON(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "\t")
}

func marshalIndentXML(v interface{}) ([]byte, error) {
	return xml.MarshalIndent(v, "", "\t")
}

func init() {
	snowWeather.StringVar(&weather.LocationName, "location", "", "Name of the location")
	snowWeather.Float64Var(&weather.Temperature, "temp", -274, "Temperature at the location in degrees Celsius")
	snowWeather.Func("loc_crd", "Coordinates information", parseLocationCoords)

	hotlyWeather.StringVar(&weather.LocationName, "location", "", "Name of the location")
	hotlyWeather.Float64Var(&weather.Temperature, "temp", -274, "Temperature at the location in degrees Celsius")
	hotlyWeather.Func("loc_crd", "Coordinates information", parseLocationCoords)
}

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	fmt.Println("\n")
	logger.Info("Вами были переданы параметры для погоды типа:", zap.String("type", os.Args[1]))

	switch os.Args[1] {
	case "snowy":
		err := snowWeather.Parse(os.Args[2:])
		if err != nil {
			logger.Fatal("Ошибка при обработке флагов:", zap.Error(err))
			return
		}
		weather.Comment = strings.Join(snowWeather.Args(), " ")
	case "hotly":
		err := hotlyWeather.Parse(os.Args[2:])
		if err != nil {
			logger.Fatal("Ошибка при обработке флагов:", zap.Error(err))
			return
		}
		weather.Comment = strings.Join(hotlyWeather.Args(), " ")
	default:
		logger.Panic("Неверный тип погоды")
		return
	}
	if weather.Temperature <= -273.15 {
		logger.Fatal("Температура не может быть ниже -273.15°! Возможно вы забыли её указать?", zap.Float64("temp", weather.Temperature))
		return
	}
	if weather.LocationCoords == [2]float64{} {
		logger.Fatal("Кажется, вы забыли указать координаты?", zap.Any("loc_crd", weather.LocationCoords))
		return
	}
	if weather.LocationName == "" {
		logger.Fatal("Кажется, вы забыли указать местоположение?", zap.Any("location", weather.LocationName))
		return
	}

	weather.LocationName = strings.ToLower(weather.LocationName)
	weather.Type = os.Args[1]
	date := time.Now()
	weather.Date = date
	sugar := logger.Sugar()

	formats := []struct {
		name    string
		ext     string
		encoder func(interface{}) ([]byte, error)
		log     string
	}{
		{"JSON", "json", marshalIndentJSON, "Go-структура \"Weather\" была сериализована в JSON-представление"},
		{"XML", "xml", marshalIndentXML, "Go-структура \"Weather\" была сериализована в XML-представление"},
		{"YAML", "yaml", yaml.Marshal, "Go-структура \"Weather\" была сериализована в YAML-представление"},
		{"TOML", "toml", toml.Marshal, "Go-структура \"Weather\" была сериализована в TOML-представление"},
	}

	for _, format := range formats {
		serResult, err := format.encoder(weather)
		if err != nil {
			sugar.Errorf("Ошибка сериализации в формат %s: %v", format.name, err)
			continue
		}

		fileName := fmt.Sprintf("weather.%s", format.ext)
		file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			sugar.Errorf("Ошибка открытия файла %s: %v", fileName, err)
			continue
		}
		defer file.Close()

		_, err = file.Write(serResult)
		if err != nil {
			sugar.Errorf("Ошибка записи в файл %s: %v", fileName, err)
			continue
		}

		sugar.Infof("%s. Файл %s был успешно обновлён", format.log, fileName)
	}
	fmt.Println("\n")
	view(weather)
}

func view(val interface{}) {
	ValInterface := reflect.ValueOf(val)
	valTypeInterface := reflect.TypeOf(val)
	if ValInterface.Kind() == reflect.Struct {
		maxWidthName := 5
		maxWidthValue := 5

		for i := 0; i < valTypeInterface.NumField(); i++ {
			if len(valTypeInterface.Field(i).Name) > maxWidthName {
				maxWidthName = len(valTypeInterface.Field(i).Name) + 5
			}
			valueStr := fmt.Sprintf("%v", ValInterface.Field(i).Interface())
			if len(valueStr) > maxWidthValue {
				maxWidthValue = len(valueStr) + 5
			}
		}

		for i := 0; i < ValInterface.NumField(); i++ {
			fieldName := valTypeInterface.Field(i).Name
			fieldTag := valTypeInterface.Field(i).Tag
			var fieldValue string

			if ValInterface.Field(i).Kind() == reflect.Pointer || ValInterface.Field(i).Kind() == reflect.Interface {
				if !ValInterface.Field(i).IsNil() {
					if ValInterface.Field(i).Elem().Kind() == reflect.Slice {
						fieldValue = fmt.Sprintf("%v", ValInterface.Field(i).Interface())
					} else {
						fieldValue = fmt.Sprintf("%v", ValInterface.Field(i).Elem().Interface())
					}

				}
			} else if ValInterface.Field(i).Kind() == reflect.Slice {
				fieldValue = fmt.Sprintf("%v", ValInterface.Field(i).Interface())
			} else {
				if !ValInterface.Field(i).IsZero() {
					fieldValue = fmt.Sprintf("%v", ValInterface.Field(i).Interface())
				}
			}

			fmt.Printf("%-*s | %-*s | %s\n", maxWidthName, fieldName, maxWidthValue, fieldValue, fieldTag)
		}

	} else {
		fmt.Printf("%v | %v", valTypeInterface, ValInterface.Interface())
	}
}
