package consumer_test

import (
	"context"
	"errors"

	"github.com/AlexTsIvanov/notification-system/cmd/notification-service/internal/consumer"
	"github.com/AlexTsIvanov/notification-system/cmd/notification-service/internal/mocks"
	"github.com/AlexTsIvanov/notification-system/pkg/types"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Consumer", func() {
	var (
		mockCtrl    *gomock.Controller
		mockReader  *mocks.MockReader
		mockFactory *mocks.MockFactory
		mockSender  *mocks.MockSender
		c           *consumer.Consumer
		ctx         context.Context
		event       types.EventContext
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockReader = mocks.NewMockReader(mockCtrl)
		mockFactory = mocks.NewMockFactory(mockCtrl)
		mockSender = mocks.NewMockSender(mockCtrl)
		c = consumer.NewConsumer(mockReader, mockFactory)
		ctx = context.TODO()
		event = types.EventContext{Payload: []byte(`{"channel":"email","content":"Test message","receiver":"test@example.com"}`)}
	})

	When("reading from the event queue fails", func() {
		BeforeEach(func() {
			mockReader.EXPECT().Read(ctx).Return(types.EventContext{}, errors.New("read error"))
		})

		It("should return an error", func() {
			err := c.HandleNotificationEvent(ctx)
			Expect(err).To(MatchError("error reading event queue: read error"))
		})
	})

	When("unmarshaling the event payload fails", func() {
		BeforeEach(func() {
			mockReader.EXPECT().Read(ctx).Return(types.EventContext{Payload: []byte(`invalid`)}, nil)
			mockReader.EXPECT().Nack(gomock.Any()).Return(nil)
		})

		It("should return an error", func() {
			err := c.HandleNotificationEvent(ctx)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("error unmarshaling event body: invalid character 'i' looking for beginning of value"))
		})
	})

	When("getting the sender from the factory fails", func() {
		BeforeEach(func() {
			mockReader.EXPECT().Read(ctx).Return(event, nil)
			mockFactory.EXPECT().GetSender("email").Return(nil, errors.New("factory error"))
			mockReader.EXPECT().Nack(gomock.Any()).Return(nil)
		})

		It("should return an error", func() {
			err := c.HandleNotificationEvent(ctx)
			Expect(err).To(MatchError("error getting channel: factory error"))
		})
	})

	When("sending the notification fails", func() {
		BeforeEach(func() {
			mockReader.EXPECT().Read(ctx).Return(event, nil)
			mockFactory.EXPECT().GetSender("email").Return(mockSender, nil)
			mockSender.EXPECT().Send("Test message", "test@example.com").Return(errors.New("send error"))
			mockReader.EXPECT().Nack(gomock.Any()).Return(nil)
		})

		It("should return an error", func() {
			err := c.HandleNotificationEvent(ctx)
			Expect(err).To(MatchError("error sending notification: send error"))
		})
	})

	When("handling the event successfully", func() {
		BeforeEach(func() {
			mockReader.EXPECT().Read(ctx).Return(event, nil)
			mockFactory.EXPECT().GetSender("email").Return(mockSender, nil)
			mockSender.EXPECT().Send("Test message", "test@example.com").Return(nil)
			mockReader.EXPECT().Ack(event).Return(nil)
		})

		It("should not return an error", func() {
			err := c.HandleNotificationEvent(ctx)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})
})
