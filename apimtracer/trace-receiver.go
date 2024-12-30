package apimtracer

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs/checkpoints"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.uber.org/zap"
)

type apimtracerReceiver struct {
	host         component.Host
	cancel       context.CancelFunc
	logger       *zap.Logger
	nextConsumer consumer.Traces
	config       *Config
}

// check example start function here:
// https://github.com/open-telemetry/opentelemetry-collector/blob/e1a688fc4ef45306b2782b68423b21334b5ebd18/receiver/otlpreceiver/otlp.go#L188

func (apimtracerRcvr *apimtracerReceiver) Start(_ context.Context, host component.Host) error {
	apimtracerRcvr.host = host
	ctx := context.Background()
	ctx, apimtracerRcvr.cancel = context.WithCancel(ctx)

	tokenCredential, _ := azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{})

	checkClient, err := container.NewClient(apimtracerRcvr.config.StorageContainerUrl, tokenCredential, &container.ClientOptions{})
	if err != nil {
		return err
	}
	// create a checkpoint store that will be used by the event hub
	checkpointStore, err := checkpoints.NewBlobStore(checkClient, nil)
	if err != nil {
		return err
	}

	consumerClient, err := azeventhubs.NewConsumerClient(apimtracerRcvr.config.FullyQualifiedNamespace, apimtracerRcvr.config.EventHubName, apimtracerRcvr.config.ConsumerGroup, tokenCredential, &azeventhubs.ConsumerClientOptions{})
	if err != nil {
		return err
	}
	defer consumerClient.Close(ctx)

	processor, err := azeventhubs.NewProcessor(consumerClient, checkpointStore, &azeventhubs.ProcessorOptions{})
	if err != nil {
		return err
	}

	//  for each partition in the event hub, create a partition client with processEvents as the function to process events
	dispatchPartitionClients := func() {
		for {
			partitionClient := processor.NextPartitionClient(context.TODO())

			if partitionClient == nil {
				break
			}

			go func() {
				if err := apimtracerRcvr.processEvents(partitionClient); err != nil {
					fmt.Printf("Trace received with body %v\n", err)
				}
			}()
		}
	}
	// run all partition clients
	go dispatchPartitionClients()

	processorCtx, processorCancel := context.WithCancel(context.TODO())
	defer processorCancel()

	if err := processor.Run(processorCtx); err != nil {
		return err
	}

	return nil
}

func (apimtracerRcvr *apimtracerReceiver) Shutdown(ctx context.Context) error {
	if apimtracerRcvr.cancel != nil {
		apimtracerRcvr.cancel()
	}
	return nil
}

func (apimtracerRcvr *apimtracerReceiver) processEvents(partitionClient *azeventhubs.ProcessorPartitionClient) error {
	defer closePartitionResources(partitionClient)
	for {
		receiveCtx, receiveCtxCancel := context.WithTimeout(context.TODO(), 1*time.Minute)
		events, err := partitionClient.ReceiveEvents(receiveCtx, 100, nil)
		receiveCtxCancel()

		if err != nil && !errors.Is(err, context.DeadlineExceeded) {
			return err
		}

		apimtracerRcvr.logger.Info(fmt.Sprintf("\nProcessing %d event(s)", len(events)))

		for _, event := range events {
			apimtracerRcvr.logger.Debug(fmt.Sprintf("Trace received with body %v\n", string(event.Body)))
			trace, err := mapToTrace(string(event.Body))
			if err != nil {
				apimtracerRcvr.logger.Error(fmt.Sprintf("Error mapping trace: %v . Ignoring the trace", err))
				continue
			}

			err = apimtracerRcvr.nextConsumer.ConsumeTraces(receiveCtx, *trace)
			if err != nil {
				return err
			}
		}

		if len(events) != 0 {
			if err := partitionClient.UpdateCheckpoint(context.TODO(), events[len(events)-1], nil); err != nil {
				return err
			}
		}
	}
}

func closePartitionResources(partitionClient *azeventhubs.ProcessorPartitionClient) {
	defer partitionClient.Close(context.TODO())
}
