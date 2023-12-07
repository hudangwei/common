package kafka

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/hudangwei/common/logger"
	"github.com/hudangwei/common/util"
	"go.uber.org/zap"
)

type KafkaConsumerFn func([]byte) error

type KafkaConsumerConfig struct {
	Version  string   `toml:"version"`
	Assignor string   `toml:"assignor"`
	Oldest   bool     `toml:"oldest"`
	Brokers  []string `toml:"brokers"`
	Group    string   `toml:"group"`
	Topics   string   `toml:"topics"`
	Worker   int      `toml:"worker"`
}

type KafkaConsumerProcess struct {
	config        *KafkaConsumerConfig
	wg            sync.WaitGroup
	ready         chan bool
	futureCh      chan []byte
	stopCh        chan struct{}
	stopOnce      sync.Once
	consumerGroup sarama.ConsumerGroup
}

func NewKafkaConsumerProcess(config *KafkaConsumerConfig) *KafkaConsumerProcess {
	if config.Worker <= 1 {
		config.Worker = 1
	}
	return &KafkaConsumerProcess{
		config:   config,
		ready:    make(chan bool),
		futureCh: make(chan []byte, config.Worker),
		stopCh:   make(chan struct{}),
	}
}

func (kcp *KafkaConsumerProcess) Start(ctx context.Context, fn KafkaConsumerFn) error {
	for i := 0; i < kcp.config.Worker; i++ {
		kcp.wg.Add(1)
		go kcp.process(fn)
	}
	config := kcp.getSaramaConfig()

	client, err := sarama.NewClient(kcp.config.Brokers, config)
	if err != nil {
		logger.Warn("new kafka client with error", zap.Error(err))
		return err
	}
	cg, err := sarama.NewConsumerGroupFromClient(kcp.config.Group, client)
	if err != nil {
		client.Close()
		logger.Warn("NewConsumerGroupFromClient with error", zap.Error(err))
		return err
	}
	kcp.consumerGroup = cg

	// client, err := sarama.NewConsumerGroup(strings.Split(kcp.config.Brokers, ","), kcp.config.Group, config)
	// if err != nil {
	// 	log.Warn("Error creating consumer group client: %v", err)
	// 	return err
	// }
	// kcp.consumerGroup = client

	kcp.wg.Add(1)
	go func() {
		defer kcp.wg.Done()
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := kcp.consumerGroup.Consume(ctx, strings.Split(kcp.config.Topics, ","), kcp); err != nil {
				logger.Warn("consumer with error", zap.Error(err))
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
			kcp.ready = make(chan bool)
		}
	}()

	<-kcp.ready // Await till the consumer has been set up
	logger.Debug("kafka consumer setup success")

	return nil
}

func (kcp *KafkaConsumerProcess) process(fn KafkaConsumerFn) {
	defer kcp.wg.Done()
	for {
		select {
		case <-kcp.stopCh:
			return
		default:
		}
		select {
		case <-kcp.stopCh:
			return
		case msg := <-kcp.futureCh:
			if msg != nil {
				util.WithRecover(func() {
					if err := fn(msg); err != nil {
						logger.Warn("kafka process with error", zap.Error(err))
					}
				})
			}
		}
	}
}

func (kcp *KafkaConsumerProcess) Stop() error {
	kcp.stopOnce.Do(func() {
		close(kcp.stopCh)
	})
	kcp.wg.Wait()
	if kcp.consumerGroup != nil {
		if err := kcp.consumerGroup.Close(); err != nil {
			logger.Warn("consumer group close with error", zap.Error(err))
		}
	}
	return nil
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (kcp *KafkaConsumerProcess) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(kcp.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (kcp *KafkaConsumerProcess) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (kcp *KafkaConsumerProcess) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/master/consumer_group.go#L27-L29
	for message := range claim.Messages() {
		// log.Printf("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
		select {
		case <-kcp.stopCh:
			return nil
		default:
		}

		// copyMsg := make([]byte, len(message.Value))
		// copy(copyMsg, message.Value)
		select {
		case <-kcp.stopCh:
			return nil
		case kcp.futureCh <- message.Value:
		}

		session.MarkMessage(message, "")
	}

	return nil
}

func (kcp *KafkaConsumerProcess) getSaramaConfig() *sarama.Config {
	version, err := sarama.ParseKafkaVersion(kcp.config.Version)
	if err != nil {
		panic(fmt.Sprintf("Error parsing Kafka version: %v", err))
	}
	config := sarama.NewConfig()
	config.Version = version

	switch kcp.config.Assignor {
	case "sticky":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky
	case "roundrobin":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	case "range":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	default:
		panic(fmt.Sprintf("Unrecognized consumer group partition assignor: %s", kcp.config.Assignor))
	}

	if kcp.config.Oldest {
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	}
	config.Metadata.RefreshFrequency = 3 * time.Minute
	config.ClientID = "security_kafka"
	return config
}
