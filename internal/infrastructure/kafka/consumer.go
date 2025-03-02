package kafka

import (
    "context"
    "encoding/json"
    "log"

    "github.com/Shopify/sarama"
    "github.com/red-velvet-workspace/banco-digital/internal/infrastructure/database"
)

type Consumer struct {
    consumer sarama.Consumer
}

func NewConsumer(brokers []string) (*Consumer, error) {
    config := sarama.NewConfig()
    config.Consumer.Return.Errors = true

    consumer, err := sarama.NewConsumer(brokers, config)
    if err != nil {
        return nil, err
    }

    return &Consumer{
        consumer: consumer,
    }, nil
}

func (c *Consumer) consumeTopic(ctx context.Context, topic string, handler func([]byte) error) error {
    partitionConsumer, err := c.consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
    if err != nil {
        log.Printf("Failed to start consumer for topic %s: %v", topic, err)
        return err
    }

    defer func() {
        if err := partitionConsumer.Close(); err != nil {
            log.Printf("Failed to close partition consumer: %v", err)
        }
    }()

    for {
        select {
        case msg := <-partitionConsumer.Messages():
            if err := handler(msg.Value); err != nil {
                log.Printf("Error processing message from topic %s: %v", topic, err)
            }
        case err := <-partitionConsumer.Errors():
            log.Printf("Error consuming message from topic %s: %v", topic, err)
            return err
        case <-ctx.Done():
            return ctx.Err()
        }
    }
}

func (c *Consumer) ConsumeAccounts() error {
    return c.consumeTopic(context.Background(), TopicAccounts, func(data []byte) error {
        var msg AccountMessage
        if err := json.Unmarshal(data, &msg); err != nil {
            return err
        }

        if msg.Operation == "CREATE" {
            return database.DB.Create(&msg.Account).Error
        }
        return database.DB.Save(&msg.Account).Error
    })
}

func (c *Consumer) ConsumePIXKeys() error {
    return c.consumeTopic(context.Background(), TopicPIX, func(data []byte) error {
        var msg PIXKeyMessage
        if err := json.Unmarshal(data, &msg); err != nil {
            return err
        }

        if msg.Operation == "CREATE" {
            return database.DB.Create(&msg.PIXKey).Error
        }
        return database.DB.Delete(&msg.PIXKey).Error
    })
}

func (c *Consumer) ConsumeCreditCards() error {
    return c.consumeTopic(context.Background(), TopicCreditCards, func(data []byte) error {
        var msg CreditCardMessage
        if err := json.Unmarshal(data, &msg); err != nil {
            return err
        }

        if msg.Operation == "CREATE" {
            return database.DB.Create(&msg.CreditCard).Error
        }
        return database.DB.Save(&msg.CreditCard).Error
    })
}

func (c *Consumer) ConsumeTransactions() error {
    return c.consumeTopic(context.Background(), TopicTransactions, func(data []byte) error {
        var msg TransactionMessage
        if err := json.Unmarshal(data, &msg); err != nil {
            return err
        }
        return database.DB.Create(&msg.Transaction).Error
    })
}

func (c *Consumer) Close() error {
    return c.consumer.Close()
}
