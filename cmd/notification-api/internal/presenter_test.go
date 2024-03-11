package notification_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	notification "github.com/AlexTsIvanov/notification-system/cmd/notification-api/internal"
	"github.com/AlexTsIvanov/notification-system/cmd/notification-api/internal/mocks"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NotificationPresenter", func() {
	var (
		mockCtrl            *gomock.Controller
		mockValidator       *mocks.MockValidator
		mockController      *mocks.MockController
		presenter           *notification.NotificationPresenter
		e                   *echo.Echo
		notificationRequest notification.NotificationRequest
		c                   echo.Context
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockValidator = mocks.NewMockValidator(mockCtrl)
		mockController = mocks.NewMockController(mockCtrl)
		presenter = notification.NewNotificationPresenter(mockController, mockValidator)
		e = echo.New()
		notificationRequest = notification.NotificationRequest{
			Channel:  "email",
			Content:  "Hello, this is a test notification!",
			Receiver: "user@example.com",
		}
	})

	When("request body is valid", func() {
		BeforeEach(func() {
			notificationRequest.Channel = ""
			requestBody, _ := json.Marshal(notificationRequest)
			req := httptest.NewRequest(http.MethodPost, "/send", bytes.NewBuffer(requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c = e.NewContext(req, httptest.NewRecorder())

			mockValidator.EXPECT().Struct(notificationRequest).Return(nil)
			mockController.EXPECT().SendNotification(gomock.Any(), gomock.Eq(notificationRequest)).Return(nil)
		})

		It("should accept the notification request", func() {
			err := presenter.HandleSendNotification(c)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.Response().Status).To(Equal(http.StatusAccepted))
		})
	})

	When("sending notification fails due to validation error", func() {
		BeforeEach(func() {
			notificationRequest.Channel = ""
			requestBody, _ := json.Marshal(notificationRequest)
			req := httptest.NewRequest(http.MethodPost, "/send", bytes.NewBuffer(requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c = e.NewContext(req, httptest.NewRecorder())

			mockValidator.EXPECT().Struct(notificationRequest).Return(errors.New("Failed to validate body"))
		})

		It("should return a validation error", func() {
			err := presenter.HandleSendNotification(c)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Failed to validate body"))
		})
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})
})
