/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package pubsub

import (
	"encoding/json"
	pubsub1 "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/common-lib/pubsub-lib/model"
	"github.com/devtron-labs/image-scanner/common"
	"github.com/devtron-labs/image-scanner/pkg/clairService"
	"go.uber.org/zap"
)

type NatSubscription interface {
	Subscribe() error
}

type NatSubscriptionImpl struct {
	pubSubClient *pubsub1.PubSubClientServiceImpl
	logger       *zap.SugaredLogger
	clairService clairService.ClairService
}

func NewNatSubscription(pubSubClient *pubsub1.PubSubClientServiceImpl,
	logger *zap.SugaredLogger,
	clairService clairService.ClairService) (*NatSubscriptionImpl, error) {
	ns := &NatSubscriptionImpl{
		pubSubClient: pubSubClient,
		logger:       logger,
		clairService: clairService,
	}
	return ns, ns.Subscribe()
}

func (impl *NatSubscriptionImpl) Subscribe() error {
	callback := func(msg *model.PubSubMsg) {
		impl.logger.Debugw("received msg", "msg", msg)
		// defer msg.Ack()
		scanConfig := &common.ImageScanEvent{}
		err := json.Unmarshal([]byte(msg.Data), scanConfig)
		if err != nil {
			impl.logger.Errorw("err in reading msg", "err", err, "msg", string(msg.Data))
			return
		}
		impl.logger.Infow("scanConfig unmarshal data", "scanConfig", scanConfig)
		// NOTE: This is not being used, thats why not updated the call
		// TODO: Will have to update if any usage in future
		// scanConfig.Image = "quay.io/coreos/clair:v2.0.0"
		_, err = impl.clairService.ScanImage(scanConfig, nil, nil)
		if err != nil {
			impl.logger.Infow("err in process msg", "err", err)
			return
		}
	}

	var loggerFunc pubsub1.LoggerFunc = func(msg model.PubSubMsg) (string, []interface{}) {
		deploymentEvent := &common.ImageScanEvent{}
		err := json.Unmarshal([]byte(msg.Data), &deploymentEvent)
		if err != nil {
			return "error while unmarshalling deploymentEvent object", []interface{}{"err", err, "msg", msg.Data}
		}
		return "got message for deployment stage completion", []interface{}{"envId", deploymentEvent.EnvId, "appId", deploymentEvent.AppId, "ciArtifactId", deploymentEvent.CiArtifactId}
	}

	err := impl.pubSubClient.Subscribe(pubsub1.TOPIC_CI_SCAN, callback, loggerFunc)
	if err != nil {
		impl.logger.Errorw("Error while subscribing to pubsub", "topic", pubsub1.TOPIC_CI_SCAN, "error", err)
	}
	return err
}
