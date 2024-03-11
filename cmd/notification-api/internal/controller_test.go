package notification_test

import (
	"context"
	"encoding/json"
	"errors"

	notification "github.com/AlexTsIvanov/notification-system/cmd/notification-api/internal"
	"github.com/AlexTsIvanov/notification-system/cmd/notification-api/internal/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NotificationController", func() {
	var (
		mockCtrl            *gomock.Controller
		mockBroker          *mocks.MockMessageBroker
		controller          *notification.NotificationController
		ctx                 context.Context
		notificationRequest notification.NotificationRequest
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockBroker = mocks.NewMockMessageBroker(mockCtrl)
		controller = notification.NewNotificationController(mockBroker)
		ctx = context.Background()
		notificationRequest = notification.NotificationRequest{
			Channel:  "email",
			Content:  "Hello, this is a test notification!",
			Receiver: "user@example.com",
		}
	})

	When("sending notification fails due to broker error", func() {
		BeforeEach(func() {
			msg, _ := json.Marshal(notificationRequest)
			mockBroker.EXPECT().Send(ctx, msg).Return(errors.New("broker error"))
		})

		It("should return a sending error", func() {
			err := controller.SendNotification(ctx, notificationRequest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("error sending notification"))
		})
	})

	When("sending notification succeeds", func() {
		BeforeEach(func() {
			msg, _ := json.Marshal(notificationRequest)
			mockBroker.EXPECT().Send(ctx, msg).Return(nil)
		})

		It("should not return an error", func() {
			err := controller.SendNotification(ctx, notificationRequest)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})
})
