package handler

import (
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"metrics-and-alerting/internal/storage/memorystorage"
	"metrics-and-alerting/pkg/logpack"
	"metrics-and-alerting/pkg/metric"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	signKey = "KeySignMetric"
)

func randFloat64() *float64 {
	rand.Seed(time.Now().UnixNano())
	val := rand.Float64()
	return &val
}

func randInt64() *int64 {
	rand.Seed(time.Now().UnixNano())
	val := rand.Int63()
	return &val
}

func NewGaugeMetric() metric.Metric {
	return metric.Metric{
		ID:    "testGauge",
		MType: metric.GaugeType,
		Value: randFloat64(),
	}
}

func NewCounterMetric() metric.Metric {
	return metric.Metric{
		ID:    "testCounter",
		MType: metric.CounterType,
		Delta: randInt64(),
	}
}

/*
// TestGetJSON - Тест на получение метрики в JSON виде
func TestGetJSON(t *testing.T) {

	cfg := config.Config{}
	cfg.DefaultConfig()
	cfg.VerifyOnUpdate = false
	cfg.SecretKey = signKey

	st := memoryStorage.InMemoryStorage{}
	errInit := st.Init(cfg)
	require.NoError(t, errInit)

	gaugeMetric := NewGaugeMetric()
	counterMetric := NewCounterMetric()

	signGauge, errSign := metric.Sign(gaugeMetric, []byte(signKey))
	require.NoError(t, errSign)
	gaugeMetric.Hash = signGauge

	signCounter, errSign := metric.Sign(counterMetric, []byte(signKey))
	require.NoError(t, errSign)
	counterMetric.Hash = signCounter

	require.NoError(t, st.Upsert(gaugeMetric))
	require.NoError(t, st.Upsert(counterMetric))

	tests := []struct {
		name            string
		acceptEncoding  string
		contentEncoding string
		method          string
		contentType     string
		wantStatus      int
		wantError       bool
		wantMetric      metric.Metric
		requestMetric   metric.Metric
	}{
		{ // Тело запроса отправляется в сжатом виде и ответ должен быть в сжатом виде
			name:            "Test get gauge -> OK",
			acceptEncoding:  GZip,
			contentEncoding: GZip,
			method:          http.MethodPost,
			contentType:     ApplicationJSON,
			wantStatus:      http.StatusOK,
			wantError:       false,
			wantMetric:      gaugeMetric,
			requestMetric: metric.Metric{
				ID:    gaugeMetric.ID,
				MType: metric.GaugeType,
			},
		},
		{ // Тело запроса отправляется без сжатия, а ответ должен быть в сжатом виде
			name:        "Test get gauge without GZIP -> OK",
			method:      http.MethodPost,
			contentType: ApplicationJSON,
			wantStatus:  http.StatusOK,
			wantError:   false,
			wantMetric:  gaugeMetric,
			requestMetric: metric.Metric{
				ID:    gaugeMetric.ID,
				MType: metric.GaugeType,
			},
		},
		{ // Запрос с некорректным HTTP методом
			name:        "Test get gauge by http.Get -> ERROR",
			method:      http.MethodGet,
			contentType: ApplicationJSON,
			wantStatus:  http.StatusMethodNotAllowed,
			wantError:   true,
			wantMetric:  gaugeMetric,
			requestMetric: metric.Metric{
				ID:    gaugeMetric.ID,
				MType: metric.GaugeType,
			},
		},
		{ // Запрос без указания заголовка Content-Type
			name:       "Test get gauge without content-type -> ERROR",
			method:     http.MethodPost,
			wantStatus: http.StatusUnsupportedMediaType,
			wantError:  true,
			wantMetric: gaugeMetric,
			requestMetric: metric.Metric{
				ID:    gaugeMetric.ID,
				MType: metric.GaugeType,
			},
		},
		{ // Тело запроса отправляется в сжатом виде и ответ должен быть в сжатом виде
			name:            "Test get counter -> OK",
			acceptEncoding:  GZip,
			contentEncoding: GZip,
			method:          http.MethodPost,
			contentType:     ApplicationJSON,
			wantStatus:      http.StatusOK,
			wantError:       false,
			wantMetric:      counterMetric,
			requestMetric: metric.Metric{
				ID:    counterMetric.ID,
				MType: metric.CounterType,
			},
		},
		{ // Тело запроса отправляется без сжатия, а ответ должен быть в сжатом виде
			name:        "Test get counter without GZIP -> OK",
			method:      http.MethodPost,
			contentType: ApplicationJSON,
			wantStatus:  http.StatusOK,
			wantError:   false,
			wantMetric:  counterMetric,
			requestMetric: metric.Metric{
				ID:    counterMetric.ID,
				MType: metric.CounterType,
			},
		},
		{ // Запрос с некорректным HTTP методом
			name:        "Test get counter by http.Get -> ERROR",
			method:      http.MethodGet,
			contentType: ApplicationJSON,
			wantStatus:  http.StatusMethodNotAllowed,
			wantError:   true,
			wantMetric:  counterMetric,
			requestMetric: metric.Metric{
				ID:    counterMetric.ID,
				MType: metric.CounterType,
			},
		},
		{ // Запрос без указания заголовка Content-Type
			name:       "Test get gauge without content-type -> ERROR",
			method:     http.MethodPost,
			wantStatus: http.StatusUnsupportedMediaType,
			wantError:  true,
			wantMetric: counterMetric,
			requestMetric: metric.Metric{
				ID:    counterMetric.ID,
				MType: metric.CounterType,
			},
		},
		{ // Запрос с неизвестным типом метрики
			name:        "Test get metric with unknown type -> ERROR",
			method:      http.MethodPost,
			contentType: ApplicationJSON,
			wantStatus:  http.StatusNotFound,
			wantError:   true,
			wantMetric:  counterMetric,
			requestMetric: metric.Metric{
				ID:    counterMetric.ID,
				MType: "ololo",
			},
		},
		{ // Запрос без указания типа метрики
			name:        "Test get metric with invalid type -> ERROR",
			method:      http.MethodPost,
			contentType: ApplicationJSON,
			wantStatus:  http.StatusNotFound,
			wantError:   true,
			wantMetric:  counterMetric,
			requestMetric: metric.Metric{
				ID: counterMetric.ID,
			},
		},
		{ // Запрос без указания названия метрики
			name:        "Test get metric with invalid id -> ERROR",
			method:      http.MethodPost,
			contentType: ApplicationJSON,
			wantStatus:  http.StatusNotFound,
			wantError:   true,
			wantMetric:  counterMetric,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encode, _ := json.Marshal(tt.requestMetric)

			nextHandler := GetJSON(&st)
			middleware := GZipHandle(nextHandler)

			var request *http.Request
			switch tt.contentEncoding {

			case GZip:
				var compress bytes.Buffer
				gzipWriter := gzip.NewWriter(&compress)
				_, errGZip := gzipWriter.Write(encode)
				require.NoError(t, errGZip)

				errGZip = gzipWriter.Close()
				require.NoError(t, errGZip)

				request = httptest.NewRequest(tt.method, PartURLValue, &compress)
				request.Header.Set(ContentEncoding, tt.contentEncoding)

			default:
				request = httptest.NewRequest(tt.method, PartURLValue, bytes.NewReader(encode))
			}

			request.Header.Set(ContentType, tt.contentType)
			request.Header.Set(AcceptEncoding, tt.acceptEncoding)

			w := httptest.NewRecorder()
			middleware.ServeHTTP(w, request)

			response := w.Result()
			defer response.Body.Close()

			require.Equal(t, tt.wantStatus, response.StatusCode)

			if !tt.wantError {
				require.Equal(t, response.Header.Get(ContentType), tt.contentType)
				require.Equal(t, response.Header.Get(ContentEncoding), tt.acceptEncoding)

				var reader io.ReadCloser
				var errReader error

				switch tt.contentEncoding {
				case GZip:
					reader, errReader = gzip.NewReader(response.Body)
				default:
					reader = response.Body
				}

				require.NoError(t, errReader)
				defer reader.Close()

				data, errData := io.ReadAll(reader)
				require.NoError(t, errData)

				var metric metric.Metric
				errDecode := json.Unmarshal(data, &metric)
				require.NoError(t, errDecode)

				assert.Equal(t, metric, tt.wantMetric)
			}
		})
	}
}
*/

/*
func TestUpdateJSON(t *testing.T) {

	cfg := config.Config{}
	cfg.DefaultConfig()
	cfg.VerifyOnUpdate = false
	cfg.SecretKey = signKey

	st := memoryStorage.InMemoryStorage{}
	errInit := st.Init(cfg)
	require.NoError(t, errInit)

	gaugeMetric := NewGaugeMetric()
	counterMetric := NewCounterMetric()

	signGauge, errSign := metric.Sign(gaugeMetric, []byte(signKey))
	require.NoError(t, errSign)
	gaugeMetric.Hash = signGauge

	signCounter, errSign := metric.Sign(counterMetric, []byte(signKey))
	require.NoError(t, errSign)
	counterMetric.Hash = signCounter

	tests := []struct {
		name            string
		contentEncoding string
		method          string
		contentType     string
		wantStatus      int
		wantError       bool
		requestMetric   metric.Metric
	}{
		{ // Запрос без указания заголовка Content-Type
			name:            "Test update without content-type -> ERROR",
			contentEncoding: GZip,
			method:          http.MethodPost,
			wantStatus:      http.StatusUnsupportedMediaType,
			wantError:       true,
			requestMetric: metric.Metric{
				ID:    "testGauge",
				MType: metric.GaugeType,
				Value: gaugeMetric.Value,
			},
		},
		{ // Запрос с некорректным HTTP методом
			name:            "Test update with http.Get -> ERROR",
			contentEncoding: GZip,
			contentType:     ApplicationJSON,
			method:          http.MethodGet,
			wantStatus:      http.StatusMethodNotAllowed,
			wantError:       true,
			requestMetric: metric.Metric{
				ID:    "testGauge",
				MType: metric.GaugeType,
				Value: gaugeMetric.Value,
				Hash:  gaugeMetric.Hash,
			},
		},
		{ // Корректный запрос обновления gauge метрики
			name:            "Test update gauge -> OK",
			contentEncoding: GZip,
			method:          http.MethodPost,
			contentType:     ApplicationJSON,
			wantStatus:      http.StatusOK,
			wantError:       false,
			requestMetric: metric.Metric{
				ID:    "testGauge",
				MType: metric.GaugeType,
				Value: gaugeMetric.Value,
				Hash:  gaugeMetric.Hash,
			},
		},
		{ // Запрос без указания значения метрики
			name:            "Test update gauge without value -> ERROR",
			contentEncoding: GZip,
			method:          http.MethodPost,
			contentType:     ApplicationJSON,
			wantStatus:      http.StatusBadRequest,
			wantError:       true,
			requestMetric: metric.Metric{
				ID:    "testGauge",
				MType: metric.GaugeType,
				Hash:  gaugeMetric.Hash,
			},
		},
		{ // Запрос без указания подписи метрики
			name:            "Test update gauge without hash -> ERROR",
			contentEncoding: GZip,
			method:          http.MethodPost,
			contentType:     ApplicationJSON,
			wantStatus:      http.StatusBadRequest,
			wantError:       true,
			requestMetric: metric.Metric{
				ID:    "testGauge",
				MType: metric.GaugeType,
			},
		},
		{ // Запрос с некорректной подписью метрики
			name:            "Test update gauge with invalid hash -> ERROR",
			contentEncoding: GZip,
			method:          http.MethodPost,
			contentType:     ApplicationJSON,
			wantStatus:      http.StatusBadRequest,
			wantError:       true,
			requestMetric: metric.Metric{
				ID:    "testGauge",
				MType: metric.GaugeType,
				Hash:  "hash_ololo",
			},
		},
		{ // Запрос метрики без типа
			name:            "Test update gauge without type -> ERROR",
			contentEncoding: GZip,
			method:          http.MethodPost,
			contentType:     ApplicationJSON,
			wantStatus:      http.StatusNotImplemented,
			wantError:       true,
			requestMetric: metric.Metric{
				ID:    "testGauge",
				Value: gaugeMetric.Value,
				Hash:  gaugeMetric.Hash,
			},
		},
		{ // Запрос метрики без названия
			name:            "Test update gauge without id -> ERROR",
			contentEncoding: GZip,
			method:          http.MethodPost,
			contentType:     ApplicationJSON,
			wantStatus:      http.StatusBadRequest,
			wantError:       true,
			requestMetric: metric.Metric{
				MType: metric.GaugeType,
				Value: gaugeMetric.Value,
				Hash:  gaugeMetric.Hash,
			},
		},
		{ // Корректный запрос обновления counter метрики с сжатием тела запроса
			name:            "Test update counter -> OK",
			contentEncoding: GZip,
			method:          http.MethodPost,
			contentType:     ApplicationJSON,
			wantStatus:      http.StatusOK,
			wantError:       false,
			requestMetric: metric.Metric{
				ID:    "testCounter",
				MType: metric.CounterType,
				Delta: counterMetric.Delta,
				Hash:  counterMetric.Hash,
			},
		},
		{ // Корректный запрос обновления counter метрики, но без сжатия тела запроса
			name:        "Test update counter (unused gzip)-> OK",
			method:      http.MethodPost,
			contentType: ApplicationJSON,
			wantStatus:  http.StatusOK,
			wantError:   false,
			requestMetric: metric.Metric{
				ID:    "testCounter",
				MType: metric.CounterType,
				Delta: counterMetric.Delta,
				Hash:  counterMetric.Hash,
			},
		},
		{ // Запрос без указания подписи метрики
			name:            "Test update counter without hash -> ERROR",
			contentEncoding: GZip,
			method:          http.MethodPost,
			contentType:     ApplicationJSON,
			wantStatus:      http.StatusBadRequest,
			wantError:       true,
			requestMetric: metric.Metric{
				ID:    "testCounter",
				MType: metric.CounterType,
			},
		},
		{ // Запрос с некорректной подписью метрики
			name:            "Test update counter with invalid hash -> ERROR",
			contentEncoding: GZip,
			method:          http.MethodPost,
			contentType:     ApplicationJSON,
			wantStatus:      http.StatusBadRequest,
			wantError:       true,
			requestMetric: metric.Metric{
				ID:    "testCounter",
				MType: metric.CounterType,
				Hash:  "ololo_hash",
			},
		},
		{ // Запрос без указания значения метрики
			name:            "Test update counter without delta -> ERROR",
			contentEncoding: GZip,
			method:          http.MethodPost,
			contentType:     ApplicationJSON,
			wantStatus:      http.StatusBadRequest,
			wantError:       true,
			requestMetric: metric.Metric{
				ID:    "testCounter",
				MType: metric.CounterType,
				Hash:  counterMetric.Hash,
			},
		},
		{ // Запрос без указания типа метрики
			name:            "Test update counter without type -> ERROR",
			contentEncoding: GZip,
			method:          http.MethodPost,
			contentType:     ApplicationJSON,
			wantStatus:      http.StatusNotImplemented,
			wantError:       true,
			requestMetric: metric.Metric{
				ID:    "testCounter",
				Delta: counterMetric.Delta,
				Hash:  counterMetric.Hash,
			},
		},
		{ // Запрос без указания названия метрики
			name:            "Test update counter without id -> ERROR",
			contentEncoding: GZip,
			method:          http.MethodPost,
			contentType:     ApplicationJSON,
			wantStatus:      http.StatusBadRequest,
			wantError:       true,
			requestMetric: metric.Metric{
				MType: metric.CounterType,
				Delta: counterMetric.Delta,
				Hash:  counterMetric.Hash,
			},
		},
	}

	for _, tt := range tests {
		errReset := st.Reset()
		require.NoError(t, errReset)

		t.Run(tt.name, func(t *testing.T) {

			encode, errEncode := json.Marshal(tt.requestMetric)
			require.NoError(t, errEncode)

			var request *http.Request
			switch tt.contentEncoding {

			case GZip:
				var compress bytes.Buffer
				gzipWriter := gzip.NewWriter(&compress)
				_, errGZip := gzipWriter.Write(encode)
				require.NoError(t, errGZip)

				errGZip = gzipWriter.Close()
				require.NoError(t, errGZip)

				request = httptest.NewRequest(tt.method, PartURLUpdate, &compress)
				request.Header.Set(ContentEncoding, tt.contentEncoding)

			default:
				request = httptest.NewRequest(tt.method, PartURLUpdate, bytes.NewReader(encode))
			}

			request.Header.Set(ContentType, tt.contentType)

			nextHandler := UpdateJSON(&st)
			middleware := GZipHandle(nextHandler)

			w := httptest.NewRecorder()
			middleware.ServeHTTP(w, request)

			response := w.Result()
			defer response.Body.Close()

			require.Equal(t, tt.wantStatus, response.StatusCode)

			if !tt.wantError {

				metricInStore := metric.Metric{
					ID:    tt.requestMetric.ID,
					MType: tt.requestMetric.MType,
				}

				var errStore error
				metricInStore, errStore = st.Get(metricInStore)
				require.NoError(t, errStore)

				switch metricInStore.MType {
				case metric.GaugeType:
					require.NotNil(t, metricInStore.Value)
					assert.Equal(t, *metricInStore.Value, *tt.requestMetric.Value)
				case metric.CounterType:
					require.NotNil(t, metricInStore.Delta)
					assert.Equal(t, *metricInStore.Delta, *tt.requestMetric.Delta)
				}
			}
		})
	}
}
*/

func TestGetMetric(t *testing.T) {

	logger := logpack.NewLogger()

	gauge, _ := metric.CreateMetric(metric.GaugeType, "testGauge", metric.WithValueFloat(100.023))
	counter, _ := metric.CreateMetric(metric.CounterType, "testCounter", metric.WithValueInt(100))

	st := memorystorage.NewStorage()
	handlers := New(st, logger)

	errUpsert := st.Upsert(gauge)
	require.NoError(t, errUpsert)

	errUpsert = st.Upsert(counter)
	require.NoError(t, errUpsert)

	type metricData struct {
		name       string
		metricType string
	}

	tests := []struct {
		name       string
		metricData metricData

		contentType string
		httpMethod  string
		wantCode    int
		wantValue   string
		wantError   bool
	}{
		{
			name: "TestGetMetric - Type {Gauge} => [OK]",
			metricData: metricData{
				name:       "testGauge",
				metricType: metric.GaugeType,
			},
			contentType: "text/plain",
			httpMethod:  http.MethodGet,
			wantCode:    http.StatusOK,
			wantValue:   "100.023",

			wantError: false,
		},
		{
			name: "TestGetMetric - Type {Gauge}, Without {Name} => [Error]",
			metricData: metricData{
				metricType: metric.GaugeType,
			},
			contentType: "text/plain",
			httpMethod:  http.MethodGet,
			wantCode:    http.StatusNotFound,

			wantError: true,
		},
		{
			name: "TestGetMetric - Without {Type} => [Error]",
			metricData: metricData{
				name: "testGauge",
			},
			contentType: "text/plain",
			httpMethod:  http.MethodGet,
			wantCode:    http.StatusNotFound,

			wantError: true,
		},
		{
			name:        "TestGetMetric - Without {Type, Name} => [Error]",
			metricData:  metricData{},
			contentType: "text/plain",
			httpMethod:  http.MethodGet,
			wantCode:    http.StatusNotFound,

			wantError: true,
		},
		{
			name: "TestGetMetric - Type {Counter} => [OK]",
			metricData: metricData{
				name:       "testCounter",
				metricType: metric.CounterType,
			},
			contentType: "text/plain",
			httpMethod:  http.MethodGet,
			wantCode:    http.StatusOK,
			wantValue:   "100",

			wantError: false,
		},
		{
			name: "TestGetMetric - Type {Counter}, Without {Name} => [Error]",
			metricData: metricData{
				metricType: metric.CounterType,
			},
			contentType: "text/plain",
			httpMethod:  http.MethodGet,
			wantCode:    http.StatusNotFound,

			wantError: true,
		},
		{
			name: "TestGetMetric - Without {Type} => [Error]",
			metricData: metricData{
				name: "testCounter",
			},
			contentType: "text/plain",
			httpMethod:  http.MethodGet,
			wantCode:    http.StatusNotFound,

			wantError: true,
		},
		{
			name:        "TestGetMetric - Without {Type, Name} => [Error]",
			metricData:  metricData{},
			contentType: "text/plain",
			httpMethod:  http.MethodGet,
			wantCode:    http.StatusNotFound,

			wantError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := "/value/"
			if len(tt.metricData.metricType) > 0 {
				target += tt.metricData.metricType
			}

			if len(tt.metricData.name) > 0 {
				if target[len(target)-1] != '/' {
					target += "/"
				}

				target += tt.metricData.name
			}

			request := httptest.NewRequest(tt.httpMethod, target, nil)
			request.Header.Set("Content-Type", tt.contentType)

			w := httptest.NewRecorder()
			h := handlers.Get()
			h.ServeHTTP(w, request)

			response := w.Result()
			defer response.Body.Close()

			require.Equal(t, tt.wantCode, response.StatusCode)

			got, err := io.ReadAll(response.Body)

			if !tt.wantError {
				require.NoError(t, err)
				require.Equal(t, tt.contentType, response.Header.Get("Content-Type"))
				assert.Equal(t, string(got), tt.wantValue)
			}

		})
	}
}

func TestUpdateMetricURL(t *testing.T) {

	logger := logpack.NewLogger()

	tests := []struct {
		name        string
		contentType string
		httpMethod  string
		metric      metric.Metric
		wantCode    int
		wantError   bool
	}{
		{
			name: "Success update metric counter -> OK",
			metric: metric.Metric{
				ID:    "testCounter",
				MType: metric.CounterType,
				Delta: randInt64(),
			},
			contentType: "text/plain",
			httpMethod:  http.MethodPost,
			wantCode:    http.StatusOK,
			wantError:   false,
		},
		{
			name: "Fail update metric counter - without id, delta -> ERROR",
			metric: metric.Metric{
				MType: metric.CounterType,
			},
			contentType: "text/plain",
			httpMethod:  http.MethodPost,
			wantCode:    http.StatusNotFound,
			wantError:   true,
		},
		{
			name: "Success update metric gauge -> OK",
			metric: metric.Metric{
				ID:    "testGauge",
				MType: metric.GaugeType,
				Value: randFloat64(),
			},
			contentType: "text/plain",
			httpMethod:  http.MethodPost,
			wantCode:    http.StatusOK,
			wantError:   false,
		},
		{
			name: "Fail update metric gauge - without id, value -> ERROR",
			metric: metric.Metric{
				MType: metric.GaugeType,
			},
			contentType: "text/plain",
			httpMethod:  http.MethodPost,
			wantCode:    http.StatusNotFound,
			wantError:   true,
		},
	}
	for _, tt := range tests {

		memoryStorage := memorystorage.NewStorage()
		handlers := New(memoryStorage, logger)

		t.Run(tt.name, func(t *testing.T) {

			target := "/update/"
			if len(tt.metric.MType) > 0 {
				target += tt.metric.MType
			}

			if len(tt.metric.ID) > 0 {
				if target[len(target)-1] != '/' {
					target += "/"
				}

				target += tt.metric.ID
			}

			if len(tt.metric.StringValue()) > 0 {
				if target[len(target)-1] != '/' {
					target += "/"
				}

				target += tt.metric.StringValue()
			}

			request := httptest.NewRequest(tt.httpMethod, target, nil)
			request.Header.Set("Content-Type", tt.contentType)

			w := httptest.NewRecorder()
			h := handlers.UpdateURL()
			h.ServeHTTP(w, request)

			response := w.Result()
			defer response.Body.Close()

			require.Equal(t, tt.wantCode, response.StatusCode)

			if !tt.wantError {
				require.Equal(t, tt.contentType, response.Header.Get("Content-Type"))

				metric, _ := metric.CreateMetric(tt.metric.MType, tt.metric.ID)
				_, err := memoryStorage.Get(metric)
				assert.NoError(t, err)
			}
		})
	}
}

/*
func TestGetMetrics(t *testing.T) {

	st := memoryStorage.InMemoryStorage{}

	gauge := metric.Metric{
		ID:    "testGauge",
		MType: metric.GaugeType,
		Value: randFloat64(),
	}

	errUpsert := st.Upsert(gauge)
	require.NoError(t, errUpsert)

	tests := []struct {
		name   string
		metric metric.Metric

		contentType string
		httpMethod  string
		wantCode    int
		wantValue   string
		wantError   bool
	}{
		{
			name: "TestGetMetrics => [OK]",
			metric: metric.Metric{
				ID:    "testGauge",
				MType: metric.GaugeType,
				Value: gauge.Value,
			},
			contentType: "text/html",
			httpMethod:  http.MethodGet,
			wantCode:    http.StatusOK,
			wantError:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.httpMethod, "/", nil)
			request.Header.Set("Content-Type", tt.contentType)

			w := httptest.NewRecorder()
			h := http.HandlerFunc(GetMetrics(&st))
			h.ServeHTTP(w, request)

			response := w.Result()
			defer response.Body.Close()

			require.Equal(t, tt.wantCode, response.StatusCode)

			got, err := io.ReadAll(response.Body)
			answer := strings.Replace(string(got), "<br/>", "", -1)

			if !tt.wantError {
				require.NoError(t, err)
				require.Equal(t, tt.contentType, response.Header.Get("Content-Type"))
				assert.Equal(t, answer, tt.metric.ShotString())
			}

		})
	}
}
*/
