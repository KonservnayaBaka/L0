package orderHandler_test

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	sht "wbL0/internal/http/handler/orderHandler"
	mocks "wbL0/internal/mocks"
	"wbL0/internal/models"
)

func TestGetOrderInfo_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		orderUID     string
		mockSetup    func(srv *mocks.OrderServiceInterface)
		expectedCode int
		expectedBody func(body []byte) error
	}{
		{
			name:     "Success - order found",
			orderUID: "order123",
			mockSetup: func(srv *mocks.OrderServiceInterface) {
				srv.On("GetOrder", mock.Anything, "order123").
					Return(&models.FullOrder{Order: models.Order{OrderUID: "order123"}}, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: func(body []byte) error {
				var resp models.OrderResponse
				if err := json.Unmarshal(body, &resp); err != nil {
					return err
				}
				if resp.OrderUID != "order123" {
					return errors.New("unexpected orderUID in response")
				}
				return nil
			},
		},
		{
			name:     "Order not found",
			orderUID: "missing_order",
			mockSetup: func(srv *mocks.OrderServiceInterface) {
				srv.On("GetOrder", mock.Anything, "missing_order").
					Return(nil, models.ErrOrderNotFound)
			},
			expectedCode: http.StatusNotFound,
			expectedBody: func(body []byte) error {
				var resp map[string]string
				if err := json.Unmarshal(body, &resp); err != nil {
					return err
				}
				if resp["error"] == "" {
					return errors.New("expected error message in response")
				}
				return nil
			},
		},
		{
			name:     "Internal server error",
			orderUID: "error_order",
			mockSetup: func(srv *mocks.OrderServiceInterface) {
				srv.On("GetOrder", mock.Anything, "error_order").
					Return(nil, errors.New("internal error"))
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: func(body []byte) error {
				var resp map[string]string
				if err := json.Unmarshal(body, &resp); err != nil {
					return err
				}
				if resp["error"] == "" {
					return errors.New("expected error message in response")
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			srvMock := &mocks.OrderServiceInterface{}
			tt.mockSetup(srvMock)
			logger := slog.Default()
			handler := sht.NewOrderHandler(srvMock, logger)
			router.GET("/order/:orderUID", handler.GetOrderInfo)

			url := "/order/" + tt.orderUID
			if tt.orderUID == "" {
				url = "/order/"
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			err := tt.expectedBody(w.Body.Bytes())
			assert.NoError(t, err)

			srvMock.AssertExpectations(t)
		})
	}
}
