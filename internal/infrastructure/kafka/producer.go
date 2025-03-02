package kafka

import (
    "encoding/json"
    "log"

    "github.com/Shopify/sarama"
)

type Producer struct {
    producer sarama.SyncProducer
}

func NewProducer(brokers []string) (*Producer, error) {
    config := sarama.NewConfig()
    config.Producer.RequiredAcks = sarama.WaitForAll
    config.Producer.Retry.Max = 5
    config.Producer.Return.Successes = true

    producer, err := sarama.NewSyncProducer(brokers, config)
    if err != nil {
        return nil, err
    }

    return &Producer{
        producer: producer,
    }, nil
}

func (p *Producer) PublishMessage(topic string, message interface{}) error {
    json, err := json.Marshal(message)
    if err != nil {
        return err
    }

    msg := &sarama.ProducerMessage{
        Topic: topic,
        Value: sarama.StringEncoder(json),
    }

    partition, offset, err := p.producer.SendMessage(msg)
    if err != nil {
        return err
    }

    log.Printf("Message published to topic %s, partition %d, offset %d\n", 
        topic, partition, offset)
    return nil
}

func (p *Producer) Close() error {
    return p.producer.Close()
}
