Чтобы работать с данным проектом вы должны:

1) Открыть терминал
2) Открыть в нём в папку проекта
3) Прописать команду формата: `go run main.go hotly/snowy -location <string> -loc_crd <[2]float64{lat ∈ [-90, 90], long ∈ [-180, 180]}> -temp <float64> <any comments you want>`

Пример: go run main.go hotly -location desert -loc_crd 20.51,-12.63 -temp 60.52 емааааан, как же тут жарко:(

После чего вы получите отчёт о погоде в форматах: json, xml, yaml и toml; так же в консоль будет выведена сериализованная структура.
